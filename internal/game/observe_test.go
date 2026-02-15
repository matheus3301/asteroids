package game

import (
	"math"
	"testing"
)

// --------------- World Accessors ---------------

func TestPlayerPosition(t *testing.T) {
	g := newPlaying()
	x, y := g.world.PlayerPosition()
	if x != ScreenWidth/2 || y != ScreenHeight/2 {
		t.Errorf("expected (%v,%v), got (%v,%v)", ScreenWidth/2.0, ScreenHeight/2.0, x, y)
	}
}

func TestPlayerVelocity(t *testing.T) {
	g := newPlaying()
	g.world.velocities[g.world.Player].X = 3.0
	g.world.velocities[g.world.Player].Y = -2.0
	vx, vy := g.world.PlayerVelocity()
	if vx != 3.0 || vy != -2.0 {
		t.Errorf("expected (3,-2), got (%v,%v)", vx, vy)
	}
}

func TestPlayerAngle(t *testing.T) {
	g := newPlaying()
	a := g.world.PlayerAngle()
	if a != -math.Pi/2 {
		t.Errorf("expected %v, got %v", -math.Pi/2, a)
	}
}

func TestPlayerInvulnerable(t *testing.T) {
	g := newPlaying()
	if !g.world.PlayerInvulnerable() {
		t.Error("player should be invulnerable after reset")
	}
	g.world.players[g.world.Player].Invulnerable = false
	if g.world.PlayerInvulnerable() {
		t.Error("player should not be invulnerable after clearing flag")
	}
}

func TestAsteroidCount(t *testing.T) {
	g := newPlaying()
	// Level 1 → 4 asteroids
	if g.world.AsteroidCount() != 4 {
		t.Errorf("expected 4 asteroids, got %d", g.world.AsteroidCount())
	}
}

func TestAsteroids_ReturnsCorrectData(t *testing.T) {
	w := NewWorld()
	e := SpawnAsteroid(w, 100, 200, SizeLarge)
	infos := w.Asteroids()
	if len(infos) != 1 {
		t.Fatalf("expected 1 asteroid, got %d", len(infos))
	}
	a := infos[0]
	if a.X != 100 || a.Y != 200 {
		t.Errorf("expected position (100,200), got (%v,%v)", a.X, a.Y)
	}
	if a.Radius != 40 {
		t.Errorf("expected radius 40, got %v", a.Radius)
	}
	if a.Size != SizeLarge {
		t.Errorf("expected SizeLarge, got %v", a.Size)
	}
	_ = e
}

func TestActiveSaucer_NilWhenNone(t *testing.T) {
	g := newPlaying()
	if g.world.ActiveSaucer() != nil {
		t.Error("expected nil saucer")
	}
}

func TestActiveSaucer_ReturnsData(t *testing.T) {
	g := newPlaying()
	s := SpawnSaucer(g.world, SaucerLarge)
	g.world.SaucerActive = s
	info := g.world.ActiveSaucer()
	if info == nil {
		t.Fatal("expected saucer info")
	}
	if info.Radius != saucerLargeRadius {
		t.Errorf("expected radius %v, got %v", saucerLargeRadius, info.Radius)
	}
	if info.Size != SaucerLarge {
		t.Errorf("expected SaucerLarge, got %v", info.Size)
	}
}

func TestSaucerBullets_Empty(t *testing.T) {
	g := newPlaying()
	if len(g.world.SaucerBullets()) != 0 {
		t.Error("expected no saucer bullets")
	}
}

func TestPlayerBullets_CountMatches(t *testing.T) {
	g := newPlaying()
	SpawnBullet(g.world, g.world.Player)
	SpawnBullet(g.world, g.world.Player)
	bullets := g.world.PlayerBullets()
	if len(bullets) != 2 {
		t.Errorf("expected 2 bullets, got %d", len(bullets))
	}
}

// --------------- AIInputSystem ---------------

func TestAIInputSystem_RotateLeft(t *testing.T) {
	w := NewWorld()
	p := SpawnPlayer(w, 400, 300)
	w.Player = p
	startAngle := w.rotations[p].Angle

	AIInputSystem(w, AIAction{RotateLeft: true})

	if w.rotations[p].Angle != startAngle-rotationSpeed {
		t.Errorf("expected angle %v, got %v", startAngle-rotationSpeed, w.rotations[p].Angle)
	}
}

func TestAIInputSystem_RotateRight(t *testing.T) {
	w := NewWorld()
	p := SpawnPlayer(w, 400, 300)
	w.Player = p
	startAngle := w.rotations[p].Angle

	AIInputSystem(w, AIAction{RotateRight: true})

	if w.rotations[p].Angle != startAngle+rotationSpeed {
		t.Errorf("expected angle %v, got %v", startAngle+rotationSpeed, w.rotations[p].Angle)
	}
}

func TestAIInputSystem_Thrust(t *testing.T) {
	w := NewWorld()
	p := SpawnPlayer(w, 400, 300)
	w.Player = p

	AIInputSystem(w, AIAction{Thrust: true})

	vel := w.velocities[p]
	if vel.X == 0 && vel.Y == 0 {
		t.Error("velocity should be non-zero after thrust")
	}
	if !w.players[p].Thrusting {
		t.Error("Thrusting flag should be set")
	}
}

func TestAIInputSystem_Shoot(t *testing.T) {
	w := NewWorld()
	p := SpawnPlayer(w, 400, 300)
	w.Player = p

	AIInputSystem(w, AIAction{Shoot: true})

	if !w.players[p].ShootPressed {
		t.Error("ShootPressed should be true")
	}
}

func TestAIInputSystem_Hyperspace(t *testing.T) {
	w := NewWorld()
	p := SpawnPlayer(w, 400, 300)
	w.Player = p

	AIInputSystem(w, AIAction{Hyperspace: true})

	if !w.players[p].HyperspacePressed {
		t.Error("HyperspacePressed should be true")
	}
}

func TestAIInputSystem_Friction(t *testing.T) {
	w := NewWorld()
	p := SpawnPlayer(w, 400, 300)
	w.Player = p
	w.velocities[p].X = 2.0
	w.velocities[p].Y = 1.0

	AIInputSystem(w, AIAction{})

	vel := w.velocities[p]
	if vel.X != 2.0*friction || vel.Y != 1.0*friction {
		t.Errorf("expected friction applied: (%v,%v), got (%v,%v)", 2.0*friction, 1.0*friction, vel.X, vel.Y)
	}
}

func TestAIInputSystem_MaxSpeedClamped(t *testing.T) {
	w := NewWorld()
	p := SpawnPlayer(w, 400, 300)
	w.Player = p
	w.rotations[p].Angle = 0 // facing right
	w.velocities[p].X = maxSpeed
	w.velocities[p].Y = 0

	AIInputSystem(w, AIAction{Thrust: true})

	vel := w.velocities[p]
	speed := math.Sqrt(vel.X*vel.X + vel.Y*vel.Y)
	// After thrust + friction, speed should not exceed maxSpeed * friction
	if speed > maxSpeed*friction+0.001 {
		t.Errorf("speed %v exceeds max %v", speed, maxSpeed*friction)
	}
}

// --------------- ExtractObservation ---------------

func TestExtractObservation_Size(t *testing.T) {
	g := newPlaying()
	obs := ExtractObservation(g.world)
	if len(obs) != ObservationSize {
		t.Errorf("expected %d elements, got %d", ObservationSize, len(obs))
	}
}

func TestExtractObservation_PlayerAtCenter(t *testing.T) {
	g := newPlaying()
	obs := ExtractObservation(g.world)
	// Player at (400,300) → normalized: (400/800)*2-1 = 0, (300/600)*2-1 = 0
	if math.Abs(obs[0]) > 0.001 || math.Abs(obs[1]) > 0.001 {
		t.Errorf("expected normalized player pos near (0,0), got (%v,%v)", obs[0], obs[1])
	}
}

func TestExtractObservation_InvulnerableFlag(t *testing.T) {
	g := newPlaying()
	obs := ExtractObservation(g.world)
	if obs[6] != 1 {
		t.Errorf("expected invulnerable=1, got %v", obs[6])
	}

	g.world.players[g.world.Player].Invulnerable = false
	obs = ExtractObservation(g.world)
	if obs[6] != 0 {
		t.Errorf("expected invulnerable=0, got %v", obs[6])
	}
}

func TestExtractObservation_BulletCount(t *testing.T) {
	g := newPlaying()
	SpawnBullet(g.world, g.world.Player)
	SpawnBullet(g.world, g.world.Player)
	obs := ExtractObservation(g.world)
	expected := 2.0 / float64(MaxPlayerBullets)
	if math.Abs(obs[7]-expected) > 0.001 {
		t.Errorf("expected bullet count %v, got %v", expected, obs[7])
	}
}

func TestExtractObservation_Lives(t *testing.T) {
	g := newPlaying()
	obs := ExtractObservation(g.world)
	expected := 3.0 / 5.0
	if math.Abs(obs[33]-expected) > 0.001 {
		t.Errorf("expected lives %v, got %v", expected, obs[33])
	}
}

func TestExtractObservation_AsteroidSlots(t *testing.T) {
	w := NewWorld()
	w.Player = SpawnPlayer(w, 400, 300)
	w.Lives = 3
	// One asteroid at known position
	SpawnAsteroid(w, 500, 300, SizeLarge)

	obs := ExtractObservation(w)

	// First asteroid slot (index 8-11) should be nonzero
	dx := obs[8]
	dy := obs[9]
	if dx == 0 && dy == 0 {
		t.Error("expected non-zero asteroid relative position")
	}
	// Remaining asteroid slots (12-27) should be zero
	for i := 12; i < 28; i++ {
		if obs[i] != 0 {
			t.Errorf("slot %d should be 0, got %v", i, obs[i])
		}
	}
}

func TestExtractObservation_SaucerPresent(t *testing.T) {
	g := newPlaying()
	s := SpawnSaucer(g.world, SaucerLarge)
	g.world.SaucerActive = s

	obs := ExtractObservation(g.world)
	if obs[28] != 1 {
		t.Errorf("expected saucer present=1, got %v", obs[28])
	}
}

func TestExtractObservation_NoSaucer(t *testing.T) {
	g := newPlaying()
	obs := ExtractObservation(g.world)
	if obs[28] != 0 {
		t.Errorf("expected saucer present=0, got %v", obs[28])
	}
	if obs[29] != 0 || obs[30] != 0 {
		t.Error("saucer deltas should be 0 when no saucer")
	}
}

func TestExtractObservation_ValuesInRange(t *testing.T) {
	g := newPlaying()
	obs := ExtractObservation(g.world)
	for i, v := range obs {
		if v < -2 || v > 2 {
			t.Errorf("obs[%d] = %v is out of reasonable range", i, v)
		}
	}
}

// --------------- wrapDelta ---------------

func TestWrapDelta_NoWrap(t *testing.T) {
	d := WrapDelta(100, 200, 800)
	if d != 100 {
		t.Errorf("expected 100, got %v", d)
	}
}

func TestWrapDelta_WrapPositive(t *testing.T) {
	// 700 to 50 on 800-wide screen: shortest path is +150 (wrap around)
	d := WrapDelta(700, 50, 800)
	if math.Abs(d-150) > 0.001 {
		t.Errorf("expected 150, got %v", d)
	}
}

func TestWrapDelta_WrapNegative(t *testing.T) {
	// 50 to 700 on 800-wide screen: shortest path is -150
	d := WrapDelta(50, 700, 800)
	if math.Abs(d-(-150)) > 0.001 {
		t.Errorf("expected -150, got %v", d)
	}
}

// --------------- Angle error & closing speed ---------------

func TestExtractObservation_AngleError(t *testing.T) {
	w := NewWorld()
	p := SpawnPlayer(w, 400, 300)
	w.Player = p
	w.Lives = 3
	// Player faces right (angle=0)
	w.rotations[p].Angle = 0
	// Asteroid directly to the right → angle error should be ~0
	SpawnAsteroid(w, 500, 300, SizeLarge)

	obs := ExtractObservation(w)
	if math.Abs(obs[34]) > 0.1 {
		t.Errorf("angle error should be near 0 when facing asteroid, got %v", obs[34])
	}
}

func TestExtractObservation_AngleError_Behind(t *testing.T) {
	w := NewWorld()
	p := SpawnPlayer(w, 400, 300)
	w.Player = p
	w.Lives = 3
	// Player faces right (angle=0)
	w.rotations[p].Angle = 0
	// Asteroid directly to the left → angle error should be near ±1 (±π)
	SpawnAsteroid(w, 300, 300, SizeLarge)

	obs := ExtractObservation(w)
	if math.Abs(obs[34]) < 0.8 {
		t.Errorf("angle error should be near ±1 when asteroid is behind, got %v", obs[34])
	}
}

func TestExtractObservation_ClosingSpeed(t *testing.T) {
	w := NewWorld()
	p := SpawnPlayer(w, 400, 300)
	w.Player = p
	w.Lives = 3
	// Moving right toward asteroid
	w.velocities[p].X = 3.0
	w.velocities[p].Y = 0
	SpawnAsteroid(w, 500, 300, SizeLarge)

	obs := ExtractObservation(w)
	if obs[35] <= 0 {
		t.Errorf("closing speed should be positive when moving toward asteroid, got %v", obs[35])
	}
}

func TestExtractObservation_ClosingSpeed_Away(t *testing.T) {
	w := NewWorld()
	p := SpawnPlayer(w, 400, 300)
	w.Player = p
	w.Lives = 3
	// Moving left, away from asteroid on the right
	w.velocities[p].X = -3.0
	w.velocities[p].Y = 0
	SpawnAsteroid(w, 500, 300, SizeLarge)

	obs := ExtractObservation(w)
	if obs[35] >= 0 {
		t.Errorf("closing speed should be negative when moving away, got %v", obs[35])
	}
}

func TestExtractObservation_NoAsteroids_NewFields(t *testing.T) {
	w := NewWorld()
	w.Player = SpawnPlayer(w, 400, 300)
	w.Lives = 3

	obs := ExtractObservation(w)
	if obs[34] != 0 || obs[35] != 0 {
		t.Errorf("angle error and closing speed should be 0 with no asteroids, got %v, %v", obs[34], obs[35])
	}
}

// --------------- SpawnInitialWave ---------------

func TestSpawnInitialWave(t *testing.T) {
	w := NewWorld()
	w.Player = SpawnPlayer(w, 400, 300)
	w.Level = 1
	SpawnInitialWave(w)
	if len(w.asteroids) != 4 {
		t.Errorf("expected 4 asteroids, got %d", len(w.asteroids))
	}
}
