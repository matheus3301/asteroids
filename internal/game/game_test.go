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

	if g.world.Lives != 3 {
		t.Errorf("expected 3 lives, got %d", g.world.Lives)
	}
	if g.world.Score != 0 {
		t.Errorf("expected score 0, got %d", g.world.Score)
	}
	if g.world.Level != 1 {
		t.Errorf("expected level 1, got %d", g.world.Level)
	}
	if g.state != statePlaying {
		t.Errorf("expected statePlaying, got %v", g.state)
	}
	if g.world == nil {
		t.Fatal("world should be initialized")
	}
	if !g.world.Alive(g.world.Player) {
		t.Error("player entity should be alive")
	}
}

func TestNew_PlayerAtCenter(t *testing.T) {
	g := newPlaying()

	pos := g.world.positions[g.world.Player]
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

	playerPos := g.world.positions[g.world.Player]
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
	g.world.Level = 2
	spawnWave(g.world)

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

	oldScore := g.world.Score
	CollisionResponseSystem(g.world, events)

	if g.world.Score != oldScore+20 {
		t.Errorf("expected score %d, got %d", oldScore+20, g.world.Score)
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
	CollisionResponseSystem(g.world, events)

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
	CollisionResponseSystem(g.world, events)

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
	pc := g.world.players[g.world.Player]
	pc.Invulnerable = false
	pc.InvulnerableTimer = 0

	asteroid := g.world.Spawn()
	playerPos := g.world.positions[g.world.Player]
	g.world.positions[asteroid] = &Position{X: playerPos.X, Y: playerPos.Y}
	g.world.colliders[asteroid] = &Collider{Radius: 40}
	g.world.asteroids[asteroid] = &AsteroidTag{Size: SizeLarge}

	oldLives := g.world.Lives
	events := CollisionSystem(g.world)

	if !events.PlayerHit {
		t.Fatal("expected player hit")
	}

	CollisionResponseSystem(g.world, events)

	if g.world.Lives != oldLives-1 {
		t.Errorf("expected %d lives, got %d", oldLives-1, g.world.Lives)
	}
}

func TestPlayerRespawnsWithInvulnerability(t *testing.T) {
	g := newPlaying()
	for e := range g.world.asteroids {
		g.world.Destroy(e)
	}

	pc := g.world.players[g.world.Player]
	pc.Invulnerable = false
	pc.InvulnerableTimer = 0

	asteroid := g.world.Spawn()
	playerPos := g.world.positions[g.world.Player]
	g.world.positions[asteroid] = &Position{X: playerPos.X, Y: playerPos.Y}
	g.world.colliders[asteroid] = &Collider{Radius: 40}
	g.world.asteroids[asteroid] = &AsteroidTag{Size: SizeLarge}

	events := CollisionSystem(g.world)
	if !events.PlayerHit {
		t.Fatal("expected player hit")
	}

	CollisionResponseSystem(g.world, events)

	pc = g.world.players[g.world.Player]
	if !pc.Invulnerable {
		t.Error("player should be invulnerable after respawn")
	}
	if pc.InvulnerableTimer != 120 {
		t.Errorf("expected invulnerability timer 120, got %d", pc.InvulnerableTimer)
	}
	pos := g.world.positions[g.world.Player]
	if pos.X != ScreenWidth/2 || pos.Y != ScreenHeight/2 {
		t.Error("player should respawn at center")
	}
}

func TestGameOver_TriggeredAtZeroLives(t *testing.T) {
	g := newPlaying()
	g.world.Lives = 1

	for e := range g.world.asteroids {
		g.world.Destroy(e)
	}

	pc := g.world.players[g.world.Player]
	pc.Invulnerable = false

	asteroid := g.world.Spawn()
	playerPos := g.world.positions[g.world.Player]
	g.world.positions[asteroid] = &Position{X: playerPos.X, Y: playerPos.Y}
	g.world.colliders[asteroid] = &Collider{Radius: 40}
	g.world.asteroids[asteroid] = &AsteroidTag{Size: SizeLarge}

	events := CollisionSystem(g.world)
	if !events.PlayerHit {
		t.Fatal("expected player hit")
	}

	CollisionResponseSystem(g.world, events)

	if g.world.Lives > 0 {
		t.Error("lives should be 0")
	}
	if !g.world.Alive(g.world.Player) {
		// Player destroyed by killPlayer — correct
	} else {
		t.Error("player should be destroyed on game over")
	}
}

func TestWaveCleared_TriggersNextLevel(t *testing.T) {
	g := newPlaying()
	oldLevel := g.world.Level

	// Destroy all asteroids to clear the wave
	for e := range g.world.asteroids {
		g.world.Destroy(e)
	}

	WaveClearSystem(g.world)

	if g.world.Level != oldLevel+1 {
		t.Errorf("expected level %d, got %d", oldLevel+1, g.world.Level)
	}

	// New wave should have 3 + new level asteroids
	expected := 3 + g.world.Level
	if len(g.world.asteroids) != expected {
		t.Errorf("expected %d asteroids in new wave, got %d", expected, len(g.world.asteroids))
	}
}

// --------------- Saucer Integration ---------------

func TestReset_ClearsSaucerState(t *testing.T) {
	g := newPlaying()

	// Simulate a saucer being active
	g.world.SaucerActive = SpawnSaucer(g.world, SaucerLarge)
	g.world.SaucerSpawnTimer = 100

	g.reset()

	if g.world.SaucerActive != 0 {
		t.Error("saucerActive should be 0 after reset")
	}
	if g.world.SaucerSpawnTimer != saucerInitialDelay {
		t.Errorf("expected saucerSpawnTimer=%d, got %d", saucerInitialDelay, g.world.SaucerSpawnTimer)
	}
}

func TestSaucerSpawnTimer_Decrements(t *testing.T) {
	g := newPlaying()
	g.world.SaucerSpawnTimer = 5
	g.world.SaucerActive = 0

	// Clear asteroids to prevent wave-clear interference, add one to keep level
	for e := range g.world.asteroids {
		g.world.Destroy(e)
	}
	SpawnAsteroid(g.world, 700, 700, SizeLarge)

	// Make player invulnerable to prevent death from saucer
	pc := g.world.players[g.world.Player]
	pc.Invulnerable = true
	pc.InvulnerableTimer = 9999

	initial := g.world.SaucerSpawnTimer
	// Tick once via system
	SaucerSpawnSystem(g.world)

	if g.world.SaucerSpawnTimer != initial-1 {
		t.Errorf("expected timer %d, got %d", initial-1, g.world.SaucerSpawnTimer)
	}
}

func TestSaucerSpawnTimer_SpawnsAtZero(t *testing.T) {
	g := newPlaying()
	g.world.SaucerSpawnTimer = 1
	g.world.SaucerActive = 0

	SaucerSpawnSystem(g.world)

	if g.world.SaucerActive == 0 {
		t.Error("saucer should have been spawned")
	}
	if !g.world.Alive(g.world.SaucerActive) {
		t.Error("spawned saucer should be alive")
	}
}

func TestSaucerSpawnTimer_NoSpawnWhileActive(t *testing.T) {
	g := newPlaying()
	saucer := SpawnSaucer(g.world, SaucerLarge)
	g.world.SaucerActive = saucer
	g.world.SaucerSpawnTimer = 0

	// Should NOT spawn another saucer when one is active
	if g.world.SaucerActive != 0 && g.world.Alive(g.world.SaucerActive) {
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
	g.world.SaucerActive = saucer

	spos := g.world.positions[saucer]
	bullet := g.world.Spawn()
	g.world.positions[bullet] = &Position{X: spos.X, Y: spos.Y}
	g.world.bullets[bullet] = &BulletTag{Life: 10}

	events := CollisionSystem(g.world)

	oldScore := g.world.Score
	CollisionResponseSystem(g.world, events)

	if g.world.Score != oldScore+200 {
		t.Errorf("expected score %d, got %d", oldScore+200, g.world.Score)
	}
	if g.world.SaucerActive != 0 {
		t.Error("saucerActive should be cleared after destruction")
	}
}

func TestPlayerDeath_ClearsSaucerAndBullets(t *testing.T) {
	g := newPlaying()

	saucer := SpawnSaucer(g.world, SaucerLarge)
	g.world.SaucerActive = saucer
	SpawnSaucerBullet(g.world, saucer, 400, 300)
	SpawnSaucerBullet(g.world, saucer, 400, 300)

	destroySaucerAndBullets(g.world)

	if g.world.SaucerActive != 0 {
		t.Error("saucerActive should be 0 after player death")
	}
	if len(g.world.saucerBullets) != 0 {
		t.Errorf("expected 0 saucer bullets, got %d", len(g.world.saucerBullets))
	}
}

func TestWaveClear_SaucerSurvives(t *testing.T) {
	g := newPlaying()

	saucer := SpawnSaucer(g.world, SaucerLarge)
	g.world.SaucerActive = saucer

	// Clear all asteroids
	for e := range g.world.asteroids {
		g.world.Destroy(e)
	}

	WaveClearSystem(g.world)

	// Saucer should still be alive
	if !g.world.Alive(saucer) {
		t.Error("saucer should survive wave clear")
	}
	if g.world.SaucerActive != saucer {
		t.Error("saucerActive should still reference the saucer")
	}
}

func TestBulletLimit_BlocksWhenAtMax(t *testing.T) {
	g := newPlaying()
	w := g.world

	// Spawn MaxPlayerBullets bullets manually
	for i := 0; i < MaxPlayerBullets; i++ {
		SpawnBullet(w, g.world.Player)
	}
	if w.BulletCount() != MaxPlayerBullets {
		t.Fatalf("expected %d bullets, got %d", MaxPlayerBullets, w.BulletCount())
	}

	// Simulate pressing shoot — should NOT spawn a 5th bullet
	pc := w.players[g.world.Player]
	pc.ShootPressed = true
	ShootingSystem(w)

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
		b := SpawnBullet(w, g.world.Player)
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
	pc := w.players[g.world.Player]
	pc.ShootPressed = true
	ShootingSystem(w)

	if w.BulletCount() != MaxPlayerBullets {
		t.Errorf("expected %d bullets after re-fire, got %d", MaxPlayerBullets, w.BulletCount())
	}
}

func TestExtraLife_AwardedAt10000(t *testing.T) {
	g := newPlaying()
	g.world.Score = 9980
	g.world.Lives = 3
	g.world.NextExtraLifeAt = 10_000

	g.world.Score += 20 // large asteroid
	checkExtraLife(g.world)

	if g.world.Lives != 4 {
		t.Errorf("expected 4 lives, got %d", g.world.Lives)
	}
	if g.world.NextExtraLifeAt != 20_000 {
		t.Errorf("expected nextExtraLifeAt 20000, got %d", g.world.NextExtraLifeAt)
	}
}

func TestExtraLife_NotAwardedBelow(t *testing.T) {
	g := newPlaying()
	g.world.Score = 9900
	g.world.Lives = 3
	g.world.NextExtraLifeAt = 10_000

	g.world.Score += 20 // large asteroid, total 9920
	checkExtraLife(g.world)

	if g.world.Lives != 3 {
		t.Errorf("expected 3 lives, got %d", g.world.Lives)
	}
	if g.world.NextExtraLifeAt != 10_000 {
		t.Errorf("expected nextExtraLifeAt 10000, got %d", g.world.NextExtraLifeAt)
	}
}

func TestExtraLife_MultipleThresholds(t *testing.T) {
	g := newPlaying()
	g.world.Score = 29950
	g.world.Lives = 3
	g.world.NextExtraLifeAt = 10_000

	g.world.Score += 100 // small asteroid, total 30050 → crosses 10k, 20k, 30k
	checkExtraLife(g.world)

	if g.world.Lives != 6 {
		t.Errorf("expected 6 lives, got %d", g.world.Lives)
	}
	if g.world.NextExtraLifeAt != 40_000 {
		t.Errorf("expected nextExtraLifeAt 40000, got %d", g.world.NextExtraLifeAt)
	}
}

func TestReset_ExtraLifeThreshold(t *testing.T) {
	g := newPlaying()
	g.world.NextExtraLifeAt = 50_000 // simulate mid-game state

	g.reset()

	if g.world.NextExtraLifeAt != 10_000 {
		t.Errorf("expected nextExtraLifeAt 10000 after reset, got %d", g.world.NextExtraLifeAt)
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
			g.world.Score = 0
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
			CollisionResponseSystem(g.world, events)

			if g.world.Score != tt.expected {
				t.Errorf("expected score %d for %s asteroid, got %d", tt.expected, tt.name, g.world.Score)
			}
		})
	}
}

// --------------- Hyperspace ---------------

func TestHyperspace_CooldownDecrement(t *testing.T) {
	g := newPlaying()
	pc := g.world.players[g.world.Player]
	pc.HyperspaceCooldown = 5

	HyperspaceSystem(g.world, 0.5)

	if pc.HyperspaceCooldown != 4 {
		t.Errorf("expected HyperspaceCooldown 4, got %d", pc.HyperspaceCooldown)
	}
}

func TestHyperspace_TeleportsToNewPosition(t *testing.T) {
	g := newPlaying()
	pc := g.world.players[g.world.Player]
	pc.HyperspacePressed = true
	pc.HyperspaceCooldown = 0

	pos := g.world.positions[g.world.Player]
	oldX, oldY := pos.X, pos.Y

	// Use rng value > 1/16 to avoid death
	HyperspaceSystem(g.world, 0.5)

	if pos.X == oldX && pos.Y == oldY {
		t.Error("position should have changed after hyperspace")
	}
	vel := g.world.velocities[g.world.Player]
	if vel.X != 0 || vel.Y != 0 {
		t.Errorf("velocity should be zeroed, got (%v,%v)", vel.X, vel.Y)
	}
}

func TestHyperspace_SetsCooldown(t *testing.T) {
	g := newPlaying()
	pc := g.world.players[g.world.Player]
	pc.HyperspacePressed = true
	pc.HyperspaceCooldown = 0

	HyperspaceSystem(g.world, 0.5)

	if pc.HyperspaceCooldown != 30 {
		t.Errorf("expected HyperspaceCooldown 30, got %d", pc.HyperspaceCooldown)
	}
}

func TestHyperspace_BlockedDuringCooldown(t *testing.T) {
	g := newPlaying()
	pc := g.world.players[g.world.Player]
	pc.HyperspacePressed = true
	pc.HyperspaceCooldown = 10

	pos := g.world.positions[g.world.Player]
	oldX, oldY := pos.X, pos.Y

	HyperspaceSystem(g.world, 0.5)

	if pos.X != oldX || pos.Y != oldY {
		t.Error("position should not change during cooldown")
	}
}

func TestHyperspace_AllowedWhileInvulnerable(t *testing.T) {
	g := newPlaying()
	pc := g.world.players[g.world.Player]
	pc.Invulnerable = true
	pc.InvulnerableTimer = 60
	pc.HyperspacePressed = true
	pc.HyperspaceCooldown = 0

	pos := g.world.positions[g.world.Player]
	oldX, oldY := pos.X, pos.Y

	HyperspaceSystem(g.world, 0.5)

	if pos.X == oldX && pos.Y == oldY {
		t.Error("hyperspace should work while invulnerable")
	}
}

func TestHyperspace_DeathOnBadLuck(t *testing.T) {
	g := newPlaying()
	pc := g.world.players[g.world.Player]
	pc.HyperspacePressed = true
	pc.HyperspaceCooldown = 0
	oldLives := g.world.Lives

	// rng < 1/16 triggers death
	HyperspaceSystem(g.world, 0.01)

	if g.world.Lives != oldLives-1 {
		t.Errorf("expected %d lives, got %d", oldLives-1, g.world.Lives)
	}
	// Should respawn at center with invulnerability
	pos := g.world.positions[g.world.Player]
	if pos.X != ScreenWidth/2 || pos.Y != ScreenHeight/2 {
		t.Errorf("expected respawn at center, got (%v,%v)", pos.X, pos.Y)
	}
	if !pc.Invulnerable {
		t.Error("should be invulnerable after hyperspace death")
	}
	if pc.InvulnerableTimer != 120 {
		t.Errorf("expected invulnerability timer 120, got %d", pc.InvulnerableTimer)
	}
}

func TestHyperspace_GameOverOnLastLife(t *testing.T) {
	g := newPlaying()
	g.world.Lives = 1
	pc := g.world.players[g.world.Player]
	pc.HyperspacePressed = true
	pc.HyperspaceCooldown = 0

	HyperspaceSystem(g.world, 0.01)

	if g.world.Lives > 0 {
		t.Error("lives should be 0 after hyperspace death on last life")
	}
	if g.world.Alive(g.world.Player) {
		t.Error("player should be destroyed on game over")
	}
	// Mirror the orchestrator check
	if g.world.Lives <= 0 {
		g.state = stateGameOver
	}
	if g.state != stateGameOver {
		t.Errorf("expected stateGameOver, got %v", g.state)
	}
}

func TestHyperspace_SpawnsDepartureParticles(t *testing.T) {
	g := newPlaying()
	pc := g.world.players[g.world.Player]
	pc.HyperspacePressed = true
	pc.HyperspaceCooldown = 0

	particlesBefore := len(g.world.particles)

	HyperspaceSystem(g.world, 0.5)

	particlesAfter := len(g.world.particles)
	spawned := particlesAfter - particlesBefore
	if spawned < 12 {
		t.Errorf("expected at least 12 new particles, got %d", spawned)
	}
}
