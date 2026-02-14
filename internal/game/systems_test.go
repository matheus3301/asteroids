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
