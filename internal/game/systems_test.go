package game

import (
	"math"
	"testing"
)

// --------------- PhysicsSystem ---------------

func TestPhysicsSystem_VelocityAppliedToPosition(t *testing.T) {
	w := NewWorld()
	e := w.Spawn()
	w.positions[e] = &Position{X: 10, Y: 20}
	w.velocities[e] = &Velocity{X: 3, Y: -5}

	PhysicsSystem(w)

	pos := w.positions[e]
	if pos.X != 13 || pos.Y != 15 {
		t.Errorf("expected (13,15), got (%v,%v)", pos.X, pos.Y)
	}
}

func TestPhysicsSystem_SpinAppliedToAngle(t *testing.T) {
	w := NewWorld()
	e := w.Spawn()
	w.positions[e] = &Position{}
	w.rotations[e] = &Rotation{Angle: 1.0, Spin: 0.5}

	PhysicsSystem(w)

	rot := w.rotations[e]
	if rot.Angle != 1.5 {
		t.Errorf("expected angle 1.5, got %v", rot.Angle)
	}
}

func TestPhysicsSystem_NoVelocityNoChange(t *testing.T) {
	w := NewWorld()
	e := w.Spawn()
	w.positions[e] = &Position{X: 5, Y: 5}
	// No velocity component attached

	PhysicsSystem(w)

	pos := w.positions[e]
	if pos.X != 5 || pos.Y != 5 {
		t.Errorf("position should not change without velocity, got (%v,%v)", pos.X, pos.Y)
	}
}

// --------------- WrapSystem ---------------

func TestWrapSystem_LeftEdge(t *testing.T) {
	w := NewWorld()
	e := w.Spawn()
	w.positions[e] = &Position{X: -1, Y: 300}
	w.wrappers[e] = true

	WrapSystem(w)

	if w.positions[e].X != ScreenWidth-1 {
		t.Errorf("expected X=%v, got %v", ScreenWidth-1, w.positions[e].X)
	}
}

func TestWrapSystem_RightEdge(t *testing.T) {
	w := NewWorld()
	e := w.Spawn()
	w.positions[e] = &Position{X: ScreenWidth + 1, Y: 300}
	w.wrappers[e] = true

	WrapSystem(w)

	if w.positions[e].X != 1 {
		t.Errorf("expected X=1, got %v", w.positions[e].X)
	}
}

func TestWrapSystem_TopEdge(t *testing.T) {
	w := NewWorld()
	e := w.Spawn()
	w.positions[e] = &Position{X: 400, Y: -1}
	w.wrappers[e] = true

	WrapSystem(w)

	if w.positions[e].Y != ScreenHeight-1 {
		t.Errorf("expected Y=%v, got %v", ScreenHeight-1, w.positions[e].Y)
	}
}

func TestWrapSystem_BottomEdge(t *testing.T) {
	w := NewWorld()
	e := w.Spawn()
	w.positions[e] = &Position{X: 400, Y: ScreenHeight + 1}
	w.wrappers[e] = true

	WrapSystem(w)

	if w.positions[e].Y != 1 {
		t.Errorf("expected Y=1, got %v", w.positions[e].Y)
	}
}

func TestWrapSystem_InsideBoundsNoOp(t *testing.T) {
	w := NewWorld()
	e := w.Spawn()
	w.positions[e] = &Position{X: 400, Y: 300}
	w.wrappers[e] = true

	WrapSystem(w)

	pos := w.positions[e]
	if pos.X != 400 || pos.Y != 300 {
		t.Errorf("position should not change inside bounds, got (%v,%v)", pos.X, pos.Y)
	}
}

func TestWrapSystem_NonWrappedEntityIgnored(t *testing.T) {
	w := NewWorld()
	e := w.Spawn()
	w.positions[e] = &Position{X: -10, Y: -10}
	// Not added to wrappers

	WrapSystem(w)

	pos := w.positions[e]
	if pos.X != -10 || pos.Y != -10 {
		t.Error("non-wrapped entity should not be affected")
	}
}

// --------------- LifetimeSystem ---------------

func TestLifetimeSystem_BulletLifeDecrements(t *testing.T) {
	w := NewWorld()
	e := w.Spawn()
	w.bullets[e] = &BulletTag{Life: 10}

	LifetimeSystem(w)

	if w.bullets[e].Life != 9 {
		t.Errorf("expected life 9, got %d", w.bullets[e].Life)
	}
}

func TestLifetimeSystem_BulletDestroyedAtZero(t *testing.T) {
	w := NewWorld()
	e := w.Spawn()
	w.bullets[e] = &BulletTag{Life: 1}

	LifetimeSystem(w)

	if w.Alive(e) {
		t.Error("bullet with life 1 should be destroyed after one tick")
	}
}

func TestLifetimeSystem_ParticleLifeDecrements(t *testing.T) {
	w := NewWorld()
	e := w.Spawn()
	w.particles[e] = &ParticleTag{Life: 15, MaxLife: 30}
	w.velocities[e] = &Velocity{X: 1, Y: 1}

	LifetimeSystem(w)

	if w.particles[e].Life != 14 {
		t.Errorf("expected particle life 14, got %d", w.particles[e].Life)
	}
}

func TestLifetimeSystem_ParticleDestroyedAtZero(t *testing.T) {
	w := NewWorld()
	e := w.Spawn()
	w.particles[e] = &ParticleTag{Life: 1, MaxLife: 10}

	LifetimeSystem(w)

	if w.Alive(e) {
		t.Error("particle with life 1 should be destroyed after one tick")
	}
}

func TestLifetimeSystem_ParticleDragApplied(t *testing.T) {
	w := NewWorld()
	e := w.Spawn()
	w.particles[e] = &ParticleTag{Life: 10, MaxLife: 10}
	w.velocities[e] = &Velocity{X: 100, Y: 100}

	LifetimeSystem(w)

	vel := w.velocities[e]
	expected := 100.0 * particleDrag
	if vel.X != expected || vel.Y != expected {
		t.Errorf("expected velocity (%v,%v), got (%v,%v)", expected, expected, vel.X, vel.Y)
	}
}

// --------------- InvulnerabilitySystem ---------------

func TestInvulnerabilitySystem_TimerCountsDown(t *testing.T) {
	w := NewWorld()
	e := w.Spawn()
	w.players[e] = &PlayerControl{
		Invulnerable:      true,
		InvulnerableTimer: 10,
		BlinkTimer:        0,
	}

	InvulnerabilitySystem(w)

	pc := w.players[e]
	if pc.InvulnerableTimer != 9 {
		t.Errorf("expected timer 9, got %d", pc.InvulnerableTimer)
	}
}

func TestInvulnerabilitySystem_BlinkTimerIncrements(t *testing.T) {
	w := NewWorld()
	e := w.Spawn()
	w.players[e] = &PlayerControl{
		Invulnerable:      true,
		InvulnerableTimer: 10,
		BlinkTimer:        5,
	}

	InvulnerabilitySystem(w)

	pc := w.players[e]
	if pc.BlinkTimer != 6 {
		t.Errorf("expected blink timer 6, got %d", pc.BlinkTimer)
	}
}

func TestInvulnerabilitySystem_ClearsWhenTimerHitsZero(t *testing.T) {
	w := NewWorld()
	e := w.Spawn()
	w.players[e] = &PlayerControl{
		Invulnerable:      true,
		InvulnerableTimer: 1,
	}

	InvulnerabilitySystem(w)

	pc := w.players[e]
	if pc.Invulnerable {
		t.Error("invulnerability should be cleared when timer reaches 0")
	}
}

func TestInvulnerabilitySystem_NoOpWhenNotInvulnerable(t *testing.T) {
	w := NewWorld()
	e := w.Spawn()
	w.players[e] = &PlayerControl{
		Invulnerable:      false,
		InvulnerableTimer: 0,
		BlinkTimer:        0,
	}

	InvulnerabilitySystem(w)

	pc := w.players[e]
	if pc.BlinkTimer != 0 {
		t.Error("blink timer should not change when not invulnerable")
	}
}

// --------------- CollisionSystem ---------------

func TestCollisionSystem_BulletAsteroidHit(t *testing.T) {
	w := NewWorld()

	bullet := w.Spawn()
	w.positions[bullet] = &Position{X: 100, Y: 100}
	w.bullets[bullet] = &BulletTag{Life: 10}

	asteroid := w.Spawn()
	w.positions[asteroid] = &Position{X: 105, Y: 100} // within radius 40
	w.colliders[asteroid] = &Collider{Radius: 40}
	w.asteroids[asteroid] = &AsteroidTag{Size: SizeLarge}

	events := CollisionSystem(w)

	if len(events.BulletHits) != 1 {
		t.Fatalf("expected 1 bullet hit, got %d", len(events.BulletHits))
	}
	if events.BulletHits[0].Bullet != bullet {
		t.Error("wrong bullet entity in hit event")
	}
	if events.BulletHits[0].Asteroid != asteroid {
		t.Error("wrong asteroid entity in hit event")
	}
}

func TestCollisionSystem_BulletAsteroidMiss(t *testing.T) {
	w := NewWorld()

	bullet := w.Spawn()
	w.positions[bullet] = &Position{X: 100, Y: 100}
	w.bullets[bullet] = &BulletTag{Life: 10}

	asteroid := w.Spawn()
	w.positions[asteroid] = &Position{X: 500, Y: 500} // far away
	w.colliders[asteroid] = &Collider{Radius: 10}
	w.asteroids[asteroid] = &AsteroidTag{Size: SizeSmall}

	events := CollisionSystem(w)

	if len(events.BulletHits) != 0 {
		t.Errorf("expected 0 bullet hits, got %d", len(events.BulletHits))
	}
}

func TestCollisionSystem_PlayerAsteroidHit(t *testing.T) {
	w := NewWorld()

	player := w.Spawn()
	w.positions[player] = &Position{X: 100, Y: 100}
	w.colliders[player] = &Collider{Radius: 15}
	w.players[player] = &PlayerControl{Invulnerable: false}

	asteroid := w.Spawn()
	w.positions[asteroid] = &Position{X: 110, Y: 100} // distance 10 < 15+40
	w.colliders[asteroid] = &Collider{Radius: 40}
	w.asteroids[asteroid] = &AsteroidTag{Size: SizeLarge}

	events := CollisionSystem(w)

	if !events.PlayerHit {
		t.Error("expected player hit")
	}
	if events.PlayerEntity != player {
		t.Error("wrong player entity in collision event")
	}
}

func TestCollisionSystem_PlayerSkippedWhenInvulnerable(t *testing.T) {
	w := NewWorld()

	player := w.Spawn()
	w.positions[player] = &Position{X: 100, Y: 100}
	w.colliders[player] = &Collider{Radius: 15}
	w.players[player] = &PlayerControl{Invulnerable: true, InvulnerableTimer: 60}

	asteroid := w.Spawn()
	w.positions[asteroid] = &Position{X: 110, Y: 100}
	w.colliders[asteroid] = &Collider{Radius: 40}
	w.asteroids[asteroid] = &AsteroidTag{Size: SizeLarge}

	events := CollisionSystem(w)

	if events.PlayerHit {
		t.Error("invulnerable player should not register a hit")
	}
}

func TestCollisionSystem_PlayerAsteroidMiss(t *testing.T) {
	w := NewWorld()

	player := w.Spawn()
	w.positions[player] = &Position{X: 100, Y: 100}
	w.colliders[player] = &Collider{Radius: 15}
	w.players[player] = &PlayerControl{Invulnerable: false}

	asteroid := w.Spawn()
	w.positions[asteroid] = &Position{X: 500, Y: 500} // far away
	w.colliders[asteroid] = &Collider{Radius: 10}
	w.asteroids[asteroid] = &AsteroidTag{Size: SizeSmall}

	events := CollisionSystem(w)

	if events.PlayerHit {
		t.Error("should not detect player hit when far apart")
	}
}

func TestCollisionSystem_ReturnsCorrectEventData(t *testing.T) {
	w := NewWorld()

	// Two bullets hitting two asteroids
	b1 := w.Spawn()
	w.positions[b1] = &Position{X: 100, Y: 100}
	w.bullets[b1] = &BulletTag{Life: 10}

	a1 := w.Spawn()
	w.positions[a1] = &Position{X: 100, Y: 100}
	w.colliders[a1] = &Collider{Radius: 40}
	w.asteroids[a1] = &AsteroidTag{Size: SizeLarge}

	b2 := w.Spawn()
	w.positions[b2] = &Position{X: 500, Y: 500}
	w.bullets[b2] = &BulletTag{Life: 10}

	a2 := w.Spawn()
	w.positions[a2] = &Position{X: 500, Y: 500}
	w.colliders[a2] = &Collider{Radius: 40}
	w.asteroids[a2] = &AsteroidTag{Size: SizeMedium}

	events := CollisionSystem(w)

	if len(events.BulletHits) != 2 {
		t.Fatalf("expected 2 bullet hits, got %d", len(events.BulletHits))
	}
}

func TestCollisionSystem_ExpiredBulletIgnored(t *testing.T) {
	w := NewWorld()

	bullet := w.Spawn()
	w.positions[bullet] = &Position{X: 100, Y: 100}
	w.bullets[bullet] = &BulletTag{Life: 0} // expired

	asteroid := w.Spawn()
	w.positions[asteroid] = &Position{X: 100, Y: 100} // overlapping
	w.colliders[asteroid] = &Collider{Radius: 40}
	w.asteroids[asteroid] = &AsteroidTag{Size: SizeLarge}

	events := CollisionSystem(w)

	if len(events.BulletHits) != 0 {
		t.Error("expired bullet should not register collision")
	}
}

func TestCollisionSystem_BulletAndPlayerHitsInSameFrame(t *testing.T) {
	w := NewWorld()

	// Bullet hitting an asteroid
	bullet := w.Spawn()
	w.positions[bullet] = &Position{X: 100, Y: 100}
	w.bullets[bullet] = &BulletTag{Life: 10}

	asteroid1 := w.Spawn()
	w.positions[asteroid1] = &Position{X: 100, Y: 100}
	w.colliders[asteroid1] = &Collider{Radius: 40}
	w.asteroids[asteroid1] = &AsteroidTag{Size: SizeLarge}

	// Player hitting a different asteroid
	player := w.Spawn()
	w.positions[player] = &Position{X: 400, Y: 400}
	w.colliders[player] = &Collider{Radius: 15}
	w.players[player] = &PlayerControl{Invulnerable: false}

	asteroid2 := w.Spawn()
	w.positions[asteroid2] = &Position{X: 400, Y: 400}
	w.colliders[asteroid2] = &Collider{Radius: 40}
	w.asteroids[asteroid2] = &AsteroidTag{Size: SizeLarge}

	events := CollisionSystem(w)

	if len(events.BulletHits) < 1 {
		t.Error("expected at least 1 bullet hit")
	}
	if !events.PlayerHit {
		t.Error("expected player hit")
	}
}

// --------------- SaucerAISystem ---------------

func TestSaucerAISystem_CooldownDecrements(t *testing.T) {
	w := NewWorld()
	e := SpawnSaucer(w, SaucerLarge)
	w.saucers[e].ShootCooldown = 10

	SaucerAISystem(w)

	if w.saucers[e].ShootCooldown >= 10 {
		t.Error("shoot cooldown should have decremented")
	}
}

func TestSaucerAISystem_FiresAtZeroCooldown(t *testing.T) {
	w := NewWorld()
	e := SpawnSaucer(w, SaucerLarge)
	w.saucers[e].ShootCooldown = 1
	w.saucers[e].VerticalTimer = 999 // prevent vertical change

	SaucerAISystem(w)

	// A saucer bullet should have been spawned
	if len(w.saucerBullets) != 1 {
		t.Errorf("expected 1 saucer bullet after fire, got %d", len(w.saucerBullets))
	}
}

func TestSaucerAISystem_CooldownResets(t *testing.T) {
	w := NewWorld()
	e := SpawnSaucer(w, SaucerLarge)
	w.saucers[e].ShootCooldown = 1
	w.saucers[e].VerticalTimer = 999

	SaucerAISystem(w)

	if w.saucers[e] == nil {
		t.Fatal("saucer should still exist")
	}
	if w.saucers[e].ShootCooldown < saucerShootCooldownMin {
		t.Errorf("cooldown should have reset to at least %d, got %d", saucerShootCooldownMin, w.saucers[e].ShootCooldown)
	}
}

func TestSaucerAISystem_VerticalTimerChange(t *testing.T) {
	w := NewWorld()
	e := SpawnSaucer(w, SaucerLarge)
	w.saucers[e].VerticalTimer = 1
	w.saucers[e].ShootCooldown = 999

	SaucerAISystem(w)

	st := w.saucers[e]
	if st == nil {
		t.Fatal("saucer should still exist")
	}
	if st.VerticalTimer < saucerVerticalTimerMin {
		t.Error("vertical timer should have reset")
	}
}

func TestSaucerAISystem_VerticalWrap(t *testing.T) {
	w := NewWorld()
	e := SpawnSaucer(w, SaucerLarge)
	w.positions[e].Y = -1
	w.saucers[e].ShootCooldown = 999
	w.saucers[e].VerticalTimer = 999

	SaucerAISystem(w)

	if w.positions[e].Y < 0 {
		t.Error("saucer Y should have wrapped")
	}
}

func TestSaucerAISystem_DespawnAtFarEdge(t *testing.T) {
	w := NewWorld()
	e := SpawnSaucer(w, SaucerLarge)
	// Force moving right and past right edge
	w.saucers[e].DirectionX = 1.0
	w.positions[e].X = ScreenWidth + 100
	w.saucers[e].ShootCooldown = 999
	w.saucers[e].VerticalTimer = 999

	SaucerAISystem(w)

	if w.Alive(e) {
		t.Error("saucer should have despawned at far edge")
	}
}

func TestSaucerAISystem_NilPlayerSafe(t *testing.T) {
	w := NewWorld()
	SpawnSaucer(w, SaucerSmall)
	// Should not panic with no player entity
	SaucerAISystem(w)
}

// --------------- SaucerBulletLifetimeSystem ---------------

func TestSaucerBulletLifetimeSystem_Decrements(t *testing.T) {
	w := NewWorld()
	e := w.Spawn()
	w.saucerBullets[e] = &SaucerBulletTag{Life: 10}

	SaucerBulletLifetimeSystem(w)

	if w.saucerBullets[e].Life != 9 {
		t.Errorf("expected life 9, got %d", w.saucerBullets[e].Life)
	}
}

func TestSaucerBulletLifetimeSystem_DestroyedAtZero(t *testing.T) {
	w := NewWorld()
	e := w.Spawn()
	w.saucerBullets[e] = &SaucerBulletTag{Life: 1}

	SaucerBulletLifetimeSystem(w)

	if w.Alive(e) {
		t.Error("saucer bullet with life 1 should be destroyed after one tick")
	}
}

// --------------- CollisionSystem (saucer extensions) ---------------

func TestCollisionSystem_BulletHitsSaucer(t *testing.T) {
	w := NewWorld()

	bullet := w.Spawn()
	w.positions[bullet] = &Position{X: 100, Y: 100}
	w.bullets[bullet] = &BulletTag{Life: 10}

	saucer := w.Spawn()
	w.positions[saucer] = &Position{X: 105, Y: 100}
	w.colliders[saucer] = &Collider{Radius: 20}
	w.saucers[saucer] = &SaucerTag{Size: SaucerLarge}

	events := CollisionSystem(w)

	if len(events.SaucerBulletHits) != 1 {
		t.Fatalf("expected 1 saucer hit, got %d", len(events.SaucerBulletHits))
	}
	if events.SaucerBulletHits[0].Bullet != bullet {
		t.Error("wrong bullet in saucer hit")
	}
	if events.SaucerBulletHits[0].Saucer != saucer {
		t.Error("wrong saucer in saucer hit")
	}
}

func TestCollisionSystem_BulletMissesSaucer(t *testing.T) {
	w := NewWorld()

	bullet := w.Spawn()
	w.positions[bullet] = &Position{X: 100, Y: 100}
	w.bullets[bullet] = &BulletTag{Life: 10}

	saucer := w.Spawn()
	w.positions[saucer] = &Position{X: 500, Y: 500}
	w.colliders[saucer] = &Collider{Radius: 20}
	w.saucers[saucer] = &SaucerTag{Size: SaucerLarge}

	events := CollisionSystem(w)

	if len(events.SaucerBulletHits) != 0 {
		t.Errorf("expected 0 saucer hits, got %d", len(events.SaucerBulletHits))
	}
}

func TestCollisionSystem_SaucerBulletHitsPlayer(t *testing.T) {
	w := NewWorld()

	player := w.Spawn()
	w.positions[player] = &Position{X: 100, Y: 100}
	w.colliders[player] = &Collider{Radius: 15}
	w.players[player] = &PlayerControl{Invulnerable: false}

	sb := w.Spawn()
	w.positions[sb] = &Position{X: 105, Y: 100} // within player radius
	w.saucerBullets[sb] = &SaucerBulletTag{Life: 10}

	events := CollisionSystem(w)

	if !events.PlayerHit {
		t.Error("saucer bullet should hit player")
	}
}

func TestCollisionSystem_SaucerBulletMissesInvulnerablePlayer(t *testing.T) {
	w := NewWorld()

	player := w.Spawn()
	w.positions[player] = &Position{X: 100, Y: 100}
	w.colliders[player] = &Collider{Radius: 15}
	w.players[player] = &PlayerControl{Invulnerable: true, InvulnerableTimer: 60}

	sb := w.Spawn()
	w.positions[sb] = &Position{X: 105, Y: 100}
	w.saucerBullets[sb] = &SaucerBulletTag{Life: 10}

	events := CollisionSystem(w)

	if events.PlayerHit {
		t.Error("saucer bullet should not hit invulnerable player")
	}
}

func TestCollisionSystem_SaucerBodyHitsPlayer(t *testing.T) {
	w := NewWorld()

	player := w.Spawn()
	w.positions[player] = &Position{X: 100, Y: 100}
	w.colliders[player] = &Collider{Radius: 15}
	w.players[player] = &PlayerControl{Invulnerable: false}

	saucer := w.Spawn()
	w.positions[saucer] = &Position{X: 110, Y: 100} // within 15+20=35
	w.colliders[saucer] = &Collider{Radius: 20}
	w.saucers[saucer] = &SaucerTag{Size: SaucerLarge}

	events := CollisionSystem(w)

	if !events.PlayerHit {
		t.Error("saucer body should hit player")
	}
}

func TestCollisionSystem_ExpiredBulletIgnoredForSaucer(t *testing.T) {
	w := NewWorld()

	bullet := w.Spawn()
	w.positions[bullet] = &Position{X: 100, Y: 100}
	w.bullets[bullet] = &BulletTag{Life: 0}

	saucer := w.Spawn()
	w.positions[saucer] = &Position{X: 100, Y: 100}
	w.colliders[saucer] = &Collider{Radius: 20}
	w.saucers[saucer] = &SaucerTag{Size: SaucerLarge}

	events := CollisionSystem(w)

	if len(events.SaucerBulletHits) != 0 {
		t.Error("expired bullet should not hit saucer")
	}
}

// --------------- PhysicsSystem subtests ---------------

func TestPhysicsSystem(t *testing.T) {
	t.Run("multiple entities updated independently", func(t *testing.T) {
		w := NewWorld()
		e1 := w.Spawn()
		w.positions[e1] = &Position{X: 0, Y: 0}
		w.velocities[e1] = &Velocity{X: 1, Y: 2}

		e2 := w.Spawn()
		w.positions[e2] = &Position{X: 10, Y: 10}
		w.velocities[e2] = &Velocity{X: -1, Y: -1}

		PhysicsSystem(w)

		if w.positions[e1].X != 1 || w.positions[e1].Y != 2 {
			t.Errorf("e1: expected (1,2), got (%v,%v)", w.positions[e1].X, w.positions[e1].Y)
		}
		if w.positions[e2].X != 9 || w.positions[e2].Y != 9 {
			t.Errorf("e2: expected (9,9), got (%v,%v)", w.positions[e2].X, w.positions[e2].Y)
		}
	})

	t.Run("rotation without position velocity", func(t *testing.T) {
		w := NewWorld()
		e := w.Spawn()
		w.positions[e] = &Position{X: 5, Y: 5}
		w.rotations[e] = &Rotation{Angle: 0, Spin: 0.1}

		PhysicsSystem(w)

		if math.Abs(w.rotations[e].Angle-0.1) > 1e-9 {
			t.Errorf("expected angle 0.1, got %v", w.rotations[e].Angle)
		}
	})
}

// --------------- New system tests ---------------

func TestShootingSystem_SpawnsBullet(t *testing.T) {
	w := NewWorld()
	e := SpawnPlayer(w, 400, 300)
	w.players[e].ShootPressed = true

	ShootingSystem(w)

	if w.BulletCount() != 1 {
		t.Errorf("expected 1 bullet, got %d", w.BulletCount())
	}
}

func TestShootingSystem_RespectsLimit(t *testing.T) {
	w := NewWorld()
	e := SpawnPlayer(w, 400, 300)
	for i := 0; i < MaxPlayerBullets; i++ {
		SpawnBullet(w, e)
	}
	w.players[e].ShootPressed = true

	ShootingSystem(w)

	if w.BulletCount() != MaxPlayerBullets {
		t.Errorf("expected %d bullets, got %d", MaxPlayerBullets, w.BulletCount())
	}
}

func TestSaucerSpawnSystem_TimerDecrement(t *testing.T) {
	w := NewWorld()
	w.SaucerActive = 0
	w.SaucerSpawnTimer = 10

	SaucerSpawnSystem(w)

	if w.SaucerSpawnTimer != 9 {
		t.Errorf("expected timer 9, got %d", w.SaucerSpawnTimer)
	}
}

func TestSaucerSpawnSystem_SpawnsAtZero(t *testing.T) {
	w := NewWorld()
	w.SaucerActive = 0
	w.SaucerSpawnTimer = 1

	SaucerSpawnSystem(w)

	if w.SaucerActive == 0 {
		t.Error("saucer should have been spawned")
	}
	if !w.Alive(w.SaucerActive) {
		t.Error("spawned saucer should be alive")
	}
}

func TestSaucerSpawnSystem_NoSpawnWhileActive(t *testing.T) {
	w := NewWorld()
	saucer := SpawnSaucer(w, SaucerLarge)
	w.SaucerActive = saucer
	w.SaucerSpawnTimer = 5

	SaucerSpawnSystem(w)

	// Timer should not have changed
	if w.SaucerSpawnTimer != 5 {
		t.Errorf("timer should not tick while saucer active, got %d", w.SaucerSpawnTimer)
	}
}

func TestSaucerDespawnSystem_DetectsDeadSaucer(t *testing.T) {
	w := NewWorld()
	saucer := SpawnSaucer(w, SaucerLarge)
	w.SaucerActive = saucer
	w.Destroy(saucer) // simulate saucer destroyed by AI despawn

	SaucerDespawnSystem(w)

	if w.SaucerActive != 0 {
		t.Error("SaucerActive should be reset to 0")
	}
	if w.SaucerSpawnTimer != saucerRespawnDelay {
		t.Errorf("expected spawn timer %d, got %d", saucerRespawnDelay, w.SaucerSpawnTimer)
	}
}

func TestCollisionResponseSystem_AsteroidScore(t *testing.T) {
	w := NewWorld()
	w.Score = 0
	w.NextExtraLifeAt = 10_000

	asteroid := w.Spawn()
	w.positions[asteroid] = &Position{X: 100, Y: 100}
	w.colliders[asteroid] = &Collider{Radius: 40}
	w.asteroids[asteroid] = &AsteroidTag{Size: SizeLarge}

	bullet := w.Spawn()
	w.positions[bullet] = &Position{X: 100, Y: 100}
	w.bullets[bullet] = &BulletTag{Life: 10}

	events := CollisionSystem(w)
	CollisionResponseSystem(w, events)

	if w.Score != 20 {
		t.Errorf("expected score 20, got %d", w.Score)
	}
}

func TestCollisionResponseSystem_AsteroidSplit(t *testing.T) {
	tests := []struct {
		name          string
		size          AsteroidSize
		expectedAfter int // number of asteroids after destruction
	}{
		{"large splits to 2 medium", SizeLarge, 2},
		{"medium splits to 2 small", SizeMedium, 2},
		{"small does not split", SizeSmall, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWorld()
			w.NextExtraLifeAt = 10_000

			asteroid := w.Spawn()
			w.positions[asteroid] = &Position{X: 100, Y: 100}
			w.colliders[asteroid] = &Collider{Radius: 40}
			w.asteroids[asteroid] = &AsteroidTag{Size: tt.size}

			bullet := w.Spawn()
			w.positions[bullet] = &Position{X: 100, Y: 100}
			w.bullets[bullet] = &BulletTag{Life: 10}

			events := CollisionSystem(w)
			CollisionResponseSystem(w, events)

			if len(w.asteroids) != tt.expectedAfter {
				t.Errorf("expected %d asteroids after split, got %d", tt.expectedAfter, len(w.asteroids))
			}
		})
	}
}

func TestCollisionResponseSystem_SaucerScore(t *testing.T) {
	tests := []struct {
		name     string
		size     SaucerSize
		expected int
	}{
		{"large saucer", SaucerLarge, 200},
		{"small saucer", SaucerSmall, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWorld()
			w.Score = 0
			w.NextExtraLifeAt = 10_000

			saucer := SpawnSaucer(w, tt.size)
			w.SaucerActive = saucer

			spos := w.positions[saucer]
			bullet := w.Spawn()
			w.positions[bullet] = &Position{X: spos.X, Y: spos.Y}
			w.bullets[bullet] = &BulletTag{Life: 10}

			events := CollisionSystem(w)
			CollisionResponseSystem(w, events)

			if w.Score != tt.expected {
				t.Errorf("expected score %d, got %d", tt.expected, w.Score)
			}
		})
	}
}

func TestCollisionResponseSystem_PlayerDeath(t *testing.T) {
	w := NewWorld()
	w.Lives = 3
	w.NextExtraLifeAt = 10_000

	player := SpawnPlayer(w, 100, 100)
	w.Player = player
	w.players[player].Invulnerable = false
	w.players[player].InvulnerableTimer = 0

	asteroid := w.Spawn()
	w.positions[asteroid] = &Position{X: 100, Y: 100}
	w.colliders[asteroid] = &Collider{Radius: 40}
	w.asteroids[asteroid] = &AsteroidTag{Size: SizeLarge}

	events := CollisionSystem(w)
	CollisionResponseSystem(w, events)

	if w.Lives != 2 {
		t.Errorf("expected 2 lives, got %d", w.Lives)
	}
	pos := w.positions[player]
	if pos.X != ScreenWidth/2 || pos.Y != ScreenHeight/2 {
		t.Errorf("expected respawn at center, got (%v,%v)", pos.X, pos.Y)
	}
	if !w.players[player].Invulnerable {
		t.Error("player should be invulnerable after respawn")
	}
}

func TestCollisionResponseSystem_PlayerDeathGameOver(t *testing.T) {
	w := NewWorld()
	w.Lives = 1
	w.NextExtraLifeAt = 10_000

	player := SpawnPlayer(w, 100, 100)
	w.Player = player
	w.players[player].Invulnerable = false
	w.players[player].InvulnerableTimer = 0

	asteroid := w.Spawn()
	w.positions[asteroid] = &Position{X: 100, Y: 100}
	w.colliders[asteroid] = &Collider{Radius: 40}
	w.asteroids[asteroid] = &AsteroidTag{Size: SizeLarge}

	events := CollisionSystem(w)
	CollisionResponseSystem(w, events)

	if w.Lives != 0 {
		t.Errorf("expected 0 lives, got %d", w.Lives)
	}
	if w.Alive(player) {
		t.Error("player should be destroyed on game over")
	}
}

func TestWaveClearSystem_SpawnsNextWave(t *testing.T) {
	w := NewWorld()
	w.Level = 1
	w.Player = SpawnPlayer(w, 400, 300)

	WaveClearSystem(w)

	if w.Level != 2 {
		t.Errorf("expected level 2, got %d", w.Level)
	}
	// 3 + level 2 = 5 asteroids
	if len(w.asteroids) != 5 {
		t.Errorf("expected 5 asteroids, got %d", len(w.asteroids))
	}
}

func TestWaveClearSystem_NoOpWithAsteroids(t *testing.T) {
	w := NewWorld()
	w.Level = 1
	SpawnAsteroid(w, 100, 100, SizeLarge)

	WaveClearSystem(w)

	if w.Level != 1 {
		t.Errorf("level should not change with asteroids present, got %d", w.Level)
	}
}

func TestKillPlayer_Respawn(t *testing.T) {
	w := NewWorld()
	w.Lives = 3
	player := SpawnPlayer(w, 100, 100)
	w.Player = player
	w.players[player].Invulnerable = false

	killPlayer(w, player)

	if w.Lives != 2 {
		t.Errorf("expected 2 lives, got %d", w.Lives)
	}
	pos := w.positions[player]
	if pos.X != ScreenWidth/2 || pos.Y != ScreenHeight/2 {
		t.Errorf("expected respawn at center, got (%v,%v)", pos.X, pos.Y)
	}
	if !w.players[player].Invulnerable {
		t.Error("player should be invulnerable after respawn")
	}
	if w.players[player].InvulnerableTimer != 120 {
		t.Errorf("expected invulnerability timer 120, got %d", w.players[player].InvulnerableTimer)
	}
}

func TestKillPlayer_GameOver(t *testing.T) {
	w := NewWorld()
	w.Lives = 1
	player := SpawnPlayer(w, 100, 100)
	w.Player = player

	killPlayer(w, player)

	if w.Lives != 0 {
		t.Errorf("expected 0 lives, got %d", w.Lives)
	}
	if w.Alive(player) {
		t.Error("player should be destroyed on game over")
	}
}

func TestRespawnPlayer_Position(t *testing.T) {
	w := NewWorld()
	player := SpawnPlayer(w, 100, 100)
	w.velocities[player].X = 5
	w.velocities[player].Y = 3
	w.rotations[player].Angle = 1.0
	w.players[player].Invulnerable = false

	respawnPlayer(w, player)

	pos := w.positions[player]
	if pos.X != ScreenWidth/2 || pos.Y != ScreenHeight/2 {
		t.Errorf("expected center position, got (%v,%v)", pos.X, pos.Y)
	}
	vel := w.velocities[player]
	if vel.X != 0 || vel.Y != 0 {
		t.Errorf("expected zero velocity, got (%v,%v)", vel.X, vel.Y)
	}
	rot := w.rotations[player]
	if rot.Angle != -math.Pi/2 {
		t.Errorf("expected angle %v, got %v", -math.Pi/2, rot.Angle)
	}
	pc := w.players[player]
	if !pc.Invulnerable {
		t.Error("expected invulnerable")
	}
	if pc.InvulnerableTimer != 120 {
		t.Errorf("expected timer 120, got %d", pc.InvulnerableTimer)
	}
}
