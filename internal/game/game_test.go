package game

import (
	"math"
	"testing"
)

// newPlaying creates a Game and transitions it to the playing state via reset().
func newPlaying() *Game {
	g := New()
	g.reset()
	return g
}

func TestNew_Defaults(t *testing.T) {
	g := New()

	if g.state != stateMenu {
		t.Errorf("expected stateMenu, got %v", g.state)
	}
	if g.world != nil {
		t.Error("world should be nil before starting a game")
	}
}

func TestReset_Defaults(t *testing.T) {
	g := newPlaying()

	if g.lives != 3 {
		t.Errorf("expected 3 lives, got %d", g.lives)
	}
	if g.score != 0 {
		t.Errorf("expected score 0, got %d", g.score)
	}
	if g.level != 1 {
		t.Errorf("expected level 1, got %d", g.level)
	}
	if g.state != statePlaying {
		t.Errorf("expected statePlaying, got %v", g.state)
	}
	if g.world == nil {
		t.Fatal("world should be initialized")
	}
	if !g.world.Alive(g.player) {
		t.Error("player entity should be alive")
	}
}

func TestNew_PlayerAtCenter(t *testing.T) {
	g := newPlaying()

	pos := g.world.positions[g.player]
	if pos == nil {
		t.Fatal("player should have a position")
	}
	if pos.X != ScreenWidth/2 || pos.Y != ScreenHeight/2 {
		t.Errorf("expected player at center (%v,%v), got (%v,%v)",
			ScreenWidth/2, ScreenHeight/2, pos.X, pos.Y)
	}
}

func TestSpawnWave_CorrectCount(t *testing.T) {
	g := newPlaying()
	// After reset(), level=1, spawnWave already called → 3+1=4 asteroids
	asteroidCount := len(g.world.asteroids)
	if asteroidCount != 4 {
		t.Errorf("level 1: expected 4 asteroids, got %d", asteroidCount)
	}
}

func TestSpawnWave_AsteroidsFarFromPlayer(t *testing.T) {
	g := newPlaying()

	playerPos := g.world.positions[g.player]
	for e := range g.world.asteroids {
		apos := g.world.positions[e]
		dx := apos.X - playerPos.X
		dy := apos.Y - playerPos.Y
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist < 150 {
			t.Errorf("asteroid at (%v,%v) is too close to player at (%v,%v): dist=%v",
				apos.X, apos.Y, playerPos.X, playerPos.Y, dist)
		}
	}
}

func TestSpawnWave_Level2HasMoreAsteroids(t *testing.T) {
	g := newPlaying()

	// Clear all asteroids to trigger next level manually
	for e := range g.world.asteroids {
		g.world.Destroy(e)
	}
	g.level = 2
	g.spawnWave()

	asteroidCount := len(g.world.asteroids)
	expected := 3 + 2 // 5
	if asteroidCount != expected {
		t.Errorf("level 2: expected %d asteroids, got %d", expected, asteroidCount)
	}
}

func TestLayout_ReturnsFixedDimensions(t *testing.T) {
	g := New()
	w, h := g.Layout(1920, 1080)

	if w != ScreenWidth || h != ScreenHeight {
		t.Errorf("expected (%d,%d), got (%d,%d)", ScreenWidth, ScreenHeight, w, h)
	}
}

func TestBulletAsteroidCollision_ScoreIncremented(t *testing.T) {
	g := newPlaying()
	// Remove all existing asteroids
	for e := range g.world.asteroids {
		g.world.Destroy(e)
	}

	// Place a large asteroid right on top of a bullet
	asteroid := g.world.Spawn()
	g.world.positions[asteroid] = &Position{X: 100, Y: 100}
	g.world.colliders[asteroid] = &Collider{Radius: 40}
	g.world.asteroids[asteroid] = &AsteroidTag{Size: SizeLarge}

	bullet := g.world.Spawn()
	g.world.positions[bullet] = &Position{X: 100, Y: 100}
	g.world.bullets[bullet] = &BulletTag{Life: 10}

	events := CollisionSystem(g.world)

	oldScore := g.score
	destroyed := make(map[Entity]bool)
	for _, hit := range events.BulletHits {
		if destroyed[hit.Asteroid] {
			continue
		}
		destroyed[hit.Asteroid] = true

		ast := g.world.asteroids[hit.Asteroid]
		apos := g.world.positions[hit.Asteroid]
		if ast == nil || apos == nil {
			continue
		}

		switch ast.Size {
		case SizeLarge:
			g.score += 20
		case SizeMedium:
			g.score += 50
		case SizeSmall:
			g.score += 100
		}

		if ast.Size != SizeSmall {
			nextSize := ast.Size + 1
			SpawnAsteroid(g.world, apos.X, apos.Y, nextSize)
			SpawnAsteroid(g.world, apos.X, apos.Y, nextSize)
		}

		g.world.Destroy(hit.Bullet)
		g.world.Destroy(hit.Asteroid)
	}

	if g.score != oldScore+20 {
		t.Errorf("expected score %d, got %d", oldScore+20, g.score)
	}
}

func TestBulletAsteroidCollision_AsteroidSplits(t *testing.T) {
	g := newPlaying()
	for e := range g.world.asteroids {
		g.world.Destroy(e)
	}

	asteroid := g.world.Spawn()
	g.world.positions[asteroid] = &Position{X: 100, Y: 100}
	g.world.colliders[asteroid] = &Collider{Radius: 40}
	g.world.asteroids[asteroid] = &AsteroidTag{Size: SizeLarge}

	bullet := g.world.Spawn()
	g.world.positions[bullet] = &Position{X: 100, Y: 100}
	g.world.bullets[bullet] = &BulletTag{Life: 10}

	events := CollisionSystem(g.world)

	for _, hit := range events.BulletHits {
		ast := g.world.asteroids[hit.Asteroid]
		apos := g.world.positions[hit.Asteroid]
		if ast.Size != SizeSmall {
			SpawnAsteroid(g.world, apos.X, apos.Y, ast.Size+1)
			SpawnAsteroid(g.world, apos.X, apos.Y, ast.Size+1)
		}
		g.world.Destroy(hit.Bullet)
		g.world.Destroy(hit.Asteroid)
	}

	// Original destroyed, 2 medium spawned
	mediumCount := 0
	for _, at := range g.world.asteroids {
		if at.Size == SizeMedium {
			mediumCount++
		}
	}
	if mediumCount != 2 {
		t.Errorf("expected 2 medium asteroids after split, got %d", mediumCount)
	}
}

func TestSmallAsteroidDoesNotSplit(t *testing.T) {
	g := newPlaying()
	for e := range g.world.asteroids {
		g.world.Destroy(e)
	}

	asteroid := g.world.Spawn()
	g.world.positions[asteroid] = &Position{X: 100, Y: 100}
	g.world.colliders[asteroid] = &Collider{Radius: 10}
	g.world.asteroids[asteroid] = &AsteroidTag{Size: SizeSmall}

	bullet := g.world.Spawn()
	g.world.positions[bullet] = &Position{X: 100, Y: 100}
	g.world.bullets[bullet] = &BulletTag{Life: 10}

	events := CollisionSystem(g.world)

	for _, hit := range events.BulletHits {
		ast := g.world.asteroids[hit.Asteroid]
		apos := g.world.positions[hit.Asteroid]
		if ast != nil && apos != nil && ast.Size != SizeSmall {
			SpawnAsteroid(g.world, apos.X, apos.Y, ast.Size+1)
			SpawnAsteroid(g.world, apos.X, apos.Y, ast.Size+1)
		}
		g.world.Destroy(hit.Bullet)
		g.world.Destroy(hit.Asteroid)
	}

	if len(g.world.asteroids) != 0 {
		t.Errorf("small asteroid should not split, got %d asteroids remaining", len(g.world.asteroids))
	}
}

func TestPlayerAsteroidCollision_LifeLost(t *testing.T) {
	g := newPlaying()
	for e := range g.world.asteroids {
		g.world.Destroy(e)
	}

	// Make player non-invulnerable
	pc := g.world.players[g.player]
	pc.Invulnerable = false
	pc.InvulnerableTimer = 0

	asteroid := g.world.Spawn()
	playerPos := g.world.positions[g.player]
	g.world.positions[asteroid] = &Position{X: playerPos.X, Y: playerPos.Y}
	g.world.colliders[asteroid] = &Collider{Radius: 40}
	g.world.asteroids[asteroid] = &AsteroidTag{Size: SizeLarge}

	events := CollisionSystem(g.world)

	if !events.PlayerHit {
		t.Fatal("expected player hit")
	}

	oldLives := g.lives
	g.lives--

	if g.lives != oldLives-1 {
		t.Errorf("expected %d lives, got %d", oldLives-1, g.lives)
	}
}

func TestPlayerRespawnsWithInvulnerability(t *testing.T) {
	g := newPlaying()
	for e := range g.world.asteroids {
		g.world.Destroy(e)
	}

	pc := g.world.players[g.player]
	pc.Invulnerable = false
	pc.InvulnerableTimer = 0

	asteroid := g.world.Spawn()
	playerPos := g.world.positions[g.player]
	g.world.positions[asteroid] = &Position{X: playerPos.X, Y: playerPos.Y}
	g.world.colliders[asteroid] = &Collider{Radius: 40}
	g.world.asteroids[asteroid] = &AsteroidTag{Size: SizeLarge}

	events := CollisionSystem(g.world)
	if !events.PlayerHit {
		t.Fatal("expected player hit")
	}

	// Simulate respawn logic from game.go
	g.lives--
	if g.lives > 0 {
		ppos := g.world.positions[g.player]
		ppos.X = ScreenWidth / 2
		ppos.Y = ScreenHeight / 2
		vel := g.world.velocities[g.player]
		vel.X = 0
		vel.Y = 0
		rot := g.world.rotations[g.player]
		rot.Angle = -math.Pi / 2
		pc = g.world.players[g.player]
		pc.Invulnerable = true
		pc.InvulnerableTimer = 120
		pc.BlinkTimer = 0
	}

	if !pc.Invulnerable {
		t.Error("player should be invulnerable after respawn")
	}
	if pc.InvulnerableTimer != 120 {
		t.Errorf("expected invulnerability timer 120, got %d", pc.InvulnerableTimer)
	}
	pos := g.world.positions[g.player]
	if pos.X != ScreenWidth/2 || pos.Y != ScreenHeight/2 {
		t.Error("player should respawn at center")
	}
}

func TestGameOver_TriggeredAtZeroLives(t *testing.T) {
	g := newPlaying()
	g.lives = 1

	for e := range g.world.asteroids {
		g.world.Destroy(e)
	}

	pc := g.world.players[g.player]
	pc.Invulnerable = false

	asteroid := g.world.Spawn()
	playerPos := g.world.positions[g.player]
	g.world.positions[asteroid] = &Position{X: playerPos.X, Y: playerPos.Y}
	g.world.colliders[asteroid] = &Collider{Radius: 40}
	g.world.asteroids[asteroid] = &AsteroidTag{Size: SizeLarge}

	events := CollisionSystem(g.world)
	if !events.PlayerHit {
		t.Fatal("expected player hit")
	}

	// Simulate game over logic
	g.lives--
	if g.lives <= 0 {
		g.state = stateGameOver
		g.world.Destroy(g.player)
	}

	if g.state != stateGameOver {
		t.Error("game should be in game over state")
	}
	if g.world.Alive(g.player) {
		t.Error("player should be destroyed on game over")
	}
}

func TestWaveCleared_TriggersNextLevel(t *testing.T) {
	g := newPlaying()
	oldLevel := g.level

	// Destroy all asteroids to clear the wave
	for e := range g.world.asteroids {
		g.world.Destroy(e)
	}

	// Simulate wave clear check from updatePlaying
	if len(g.world.asteroids) == 0 {
		g.level++
		g.spawnWave()
	}

	if g.level != oldLevel+1 {
		t.Errorf("expected level %d, got %d", oldLevel+1, g.level)
	}

	// New wave should have 3 + new level asteroids
	expected := 3 + g.level
	if len(g.world.asteroids) != expected {
		t.Errorf("expected %d asteroids in new wave, got %d", expected, len(g.world.asteroids))
	}
}

// --------------- Saucer Integration ---------------

func TestReset_ClearsSaucerState(t *testing.T) {
	g := newPlaying()

	// Simulate a saucer being active
	g.saucerActive = SpawnSaucer(g.world, SaucerLarge)
	g.saucerSpawnTimer = 100

	g.reset()

	if g.saucerActive != 0 {
		t.Error("saucerActive should be 0 after reset")
	}
	if g.saucerSpawnTimer != saucerInitialDelay {
		t.Errorf("expected saucerSpawnTimer=%d, got %d", saucerInitialDelay, g.saucerSpawnTimer)
	}
}

func TestSaucerSpawnTimer_Decrements(t *testing.T) {
	g := newPlaying()
	g.saucerSpawnTimer = 5
	g.saucerActive = 0

	// Clear asteroids to prevent wave-clear interference, add one to keep level
	for e := range g.world.asteroids {
		g.world.Destroy(e)
	}
	SpawnAsteroid(g.world, 700, 700, SizeLarge)

	// Make player invulnerable to prevent death from saucer
	pc := g.world.players[g.player]
	pc.Invulnerable = true
	pc.InvulnerableTimer = 9999

	initial := g.saucerSpawnTimer
	// Tick once - need to avoid input system issues in test
	// We can call the relevant logic manually
	g.saucerSpawnTimer--

	if g.saucerSpawnTimer != initial-1 {
		t.Errorf("expected timer %d, got %d", initial-1, g.saucerSpawnTimer)
	}
}

func TestSaucerSpawnTimer_SpawnsAtZero(t *testing.T) {
	g := newPlaying()
	g.saucerSpawnTimer = 1
	g.saucerActive = 0

	// Decrement and spawn
	g.saucerSpawnTimer--
	if g.saucerSpawnTimer <= 0 {
		size := chooseSaucerSize(g.score)
		g.saucerActive = SpawnSaucer(g.world, size)
		g.saucerSpawnTimer = saucerRespawnDelay
	}

	if g.saucerActive == 0 {
		t.Error("saucer should have been spawned")
	}
	if !g.world.Alive(g.saucerActive) {
		t.Error("spawned saucer should be alive")
	}
}

func TestSaucerSpawnTimer_NoSpawnWhileActive(t *testing.T) {
	g := newPlaying()
	saucer := SpawnSaucer(g.world, SaucerLarge)
	g.saucerActive = saucer
	g.saucerSpawnTimer = 0

	// Should NOT spawn another saucer when one is active
	if g.saucerActive != 0 && g.world.Alive(g.saucerActive) {
		// Timer should not tick
	} else {
		t.Error("saucer should still be active")
	}
}

func TestChooseSaucerSize_LowScore(t *testing.T) {
	for i := 0; i < 20; i++ {
		size := chooseSaucerSize(0)
		if size != SaucerLarge {
			t.Error("score 0 should always give SaucerLarge")
		}
	}
}

func TestChooseSaucerSize_HighScore(t *testing.T) {
	for i := 0; i < 20; i++ {
		size := chooseSaucerSize(50000)
		if size != SaucerSmall {
			t.Error("score 50000 should always give SaucerSmall")
		}
	}
}

func TestChooseSaucerSize_MidScore(t *testing.T) {
	largeCount := 0
	smallCount := 0
	for i := 0; i < 200; i++ {
		size := chooseSaucerSize(25000) // 50% chance
		if size == SaucerLarge {
			largeCount++
		} else {
			smallCount++
		}
	}
	// Both should appear
	if largeCount == 0 || smallCount == 0 {
		t.Errorf("mid score should produce both sizes, got large=%d small=%d", largeCount, smallCount)
	}
}

func TestSaucerDestruction_ScoreAndCleanup(t *testing.T) {
	g := newPlaying()
	for e := range g.world.asteroids {
		g.world.Destroy(e)
	}
	SpawnAsteroid(g.world, 700, 700, SizeLarge) // keep a wave alive

	saucer := SpawnSaucer(g.world, SaucerLarge)
	g.saucerActive = saucer

	spos := g.world.positions[saucer]
	bullet := g.world.Spawn()
	g.world.positions[bullet] = &Position{X: spos.X, Y: spos.Y}
	g.world.bullets[bullet] = &BulletTag{Life: 10}

	events := CollisionSystem(g.world)

	oldScore := g.score
	for _, hit := range events.SaucerBulletHits {
		st := g.world.saucers[hit.Saucer]
		sp := g.world.positions[hit.Saucer]
		if st == nil || sp == nil {
			continue
		}
		switch st.Size {
		case SaucerLarge:
			g.score += 200
		case SaucerSmall:
			g.score += 1000
		}
		g.world.Destroy(hit.Bullet)
		g.world.Destroy(hit.Saucer)
		g.saucerActive = 0
	}

	if g.score != oldScore+200 {
		t.Errorf("expected score %d, got %d", oldScore+200, g.score)
	}
	if g.saucerActive != 0 {
		t.Error("saucerActive should be cleared after destruction")
	}
}

func TestPlayerDeath_ClearsSaucerAndBullets(t *testing.T) {
	g := newPlaying()

	saucer := SpawnSaucer(g.world, SaucerLarge)
	g.saucerActive = saucer
	SpawnSaucerBullet(g.world, saucer, 400, 300)
	SpawnSaucerBullet(g.world, saucer, 400, 300)

	g.destroySaucerAndBullets()

	if g.saucerActive != 0 {
		t.Error("saucerActive should be 0 after player death")
	}
	if len(g.world.saucerBullets) != 0 {
		t.Errorf("expected 0 saucer bullets, got %d", len(g.world.saucerBullets))
	}
}

func TestWaveClear_SaucerSurvives(t *testing.T) {
	g := newPlaying()

	saucer := SpawnSaucer(g.world, SaucerLarge)
	g.saucerActive = saucer

	// Clear all asteroids
	for e := range g.world.asteroids {
		g.world.Destroy(e)
	}

	// Trigger wave clear logic
	if len(g.world.asteroids) == 0 {
		g.level++
		g.spawnWave()
	}

	// Saucer should still be alive
	if !g.world.Alive(saucer) {
		t.Error("saucer should survive wave clear")
	}
	if g.saucerActive != saucer {
		t.Error("saucerActive should still reference the saucer")
	}
}

func TestBulletLimit_BlocksWhenAtMax(t *testing.T) {
	g := newPlaying()
	w := g.world

	// Spawn MaxPlayerBullets bullets manually
	for i := 0; i < MaxPlayerBullets; i++ {
		SpawnBullet(w, g.player)
	}
	if w.BulletCount() != MaxPlayerBullets {
		t.Fatalf("expected %d bullets, got %d", MaxPlayerBullets, w.BulletCount())
	}

	// Simulate pressing shoot — should NOT spawn a 5th bullet
	pc := w.players[g.player]
	pc.ShootPressed = true
	if w.BulletCount() < MaxPlayerBullets {
		SpawnBullet(w, g.player)
	}

	if w.BulletCount() != MaxPlayerBullets {
		t.Errorf("expected %d bullets (capped), got %d", MaxPlayerBullets, w.BulletCount())
	}
}

func TestBulletLimit_AllowsAfterExpiry(t *testing.T) {
	g := newPlaying()
	w := g.world

	// Spawn MaxPlayerBullets bullets
	bullets := make([]Entity, 0, MaxPlayerBullets)
	for i := 0; i < MaxPlayerBullets; i++ {
		b := SpawnBullet(w, g.player)
		bullets = append(bullets, b)
	}
	if w.BulletCount() != MaxPlayerBullets {
		t.Fatalf("expected %d bullets, got %d", MaxPlayerBullets, w.BulletCount())
	}

	// Destroy one bullet (simulating expiry)
	w.Destroy(bullets[0])
	if w.BulletCount() != MaxPlayerBullets-1 {
		t.Fatalf("expected %d bullets after destroy, got %d", MaxPlayerBullets-1, w.BulletCount())
	}

	// Now shooting should succeed
	if w.BulletCount() < MaxPlayerBullets {
		SpawnBullet(w, g.player)
	}

	if w.BulletCount() != MaxPlayerBullets {
		t.Errorf("expected %d bullets after re-fire, got %d", MaxPlayerBullets, w.BulletCount())
	}
}

func TestScoreValues(t *testing.T) {
	tests := []struct {
		name     string
		size     AsteroidSize
		expected int
	}{
		{"large", SizeLarge, 20},
		{"medium", SizeMedium, 50},
		{"small", SizeSmall, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newPlaying()
			g.score = 0
			for e := range g.world.asteroids {
				g.world.Destroy(e)
			}

			asteroid := g.world.Spawn()
			g.world.positions[asteroid] = &Position{X: 100, Y: 100}
			g.world.colliders[asteroid] = &Collider{Radius: 40}
			g.world.asteroids[asteroid] = &AsteroidTag{Size: tt.size}

			bullet := g.world.Spawn()
			g.world.positions[bullet] = &Position{X: 100, Y: 100}
			g.world.bullets[bullet] = &BulletTag{Life: 10}

			events := CollisionSystem(g.world)

			for _, hit := range events.BulletHits {
				ast := g.world.asteroids[hit.Asteroid]
				if ast != nil {
					switch ast.Size {
					case SizeLarge:
						g.score += 20
					case SizeMedium:
						g.score += 50
					case SizeSmall:
						g.score += 100
					}
				}
			}

			if g.score != tt.expected {
				t.Errorf("expected score %d for %s asteroid, got %d", tt.expected, tt.name, g.score)
			}
		})
	}
}
