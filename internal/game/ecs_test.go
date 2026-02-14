package game

import "testing"

func TestSpawnReturnsIncrementingIDs(t *testing.T) {
	w := NewWorld()
	e1 := w.Spawn()
	e2 := w.Spawn()
	e3 := w.Spawn()

	if e1 != 1 || e2 != 2 || e3 != 3 {
		t.Errorf("expected IDs 1,2,3 got %d,%d,%d", e1, e2, e3)
	}
}

func TestDestroyRemovesFromAllStores(t *testing.T) {
	w := NewWorld()
	e := w.Spawn()

	// Attach every component type
	w.positions[e] = &Position{X: 1, Y: 2}
	w.velocities[e] = &Velocity{X: 3, Y: 4}
	w.rotations[e] = &Rotation{Angle: 0.5}
	w.colliders[e] = &Collider{Radius: 10}
	w.renderables[e] = &Renderable{}
	w.players[e] = &PlayerControl{}
	w.asteroids[e] = &AsteroidTag{}
	w.bullets[e] = &BulletTag{Life: 10}
	w.particles[e] = &ParticleTag{Life: 5, MaxLife: 5}
	w.wrappers[e] = true

	w.Destroy(e)

	if w.Alive(e) {
		t.Fatal("entity should not be alive after Destroy")
	}
	if w.positions[e] != nil {
		t.Error("positions not cleaned up")
	}
	if w.velocities[e] != nil {
		t.Error("velocities not cleaned up")
	}
	if w.rotations[e] != nil {
		t.Error("rotations not cleaned up")
	}
	if w.colliders[e] != nil {
		t.Error("colliders not cleaned up")
	}
	if w.renderables[e] != nil {
		t.Error("renderables not cleaned up")
	}
	if w.players[e] != nil {
		t.Error("players not cleaned up")
	}
	if w.asteroids[e] != nil {
		t.Error("asteroids not cleaned up")
	}
	if w.bullets[e] != nil {
		t.Error("bullets not cleaned up")
	}
	if w.particles[e] != nil {
		t.Error("particles not cleaned up")
	}
	if w.wrappers[e] {
		t.Error("wrappers not cleaned up")
	}
}

func TestAliveBeforeAndAfterDestroy(t *testing.T) {
	w := NewWorld()
	e := w.Spawn()

	if !w.Alive(e) {
		t.Fatal("newly spawned entity should be alive")
	}

	w.Destroy(e)

	if w.Alive(e) {
		t.Fatal("destroyed entity should not be alive")
	}
}

func TestSpawnAfterDestroyDoesNotReuseIDs(t *testing.T) {
	w := NewWorld()
	e1 := w.Spawn() // ID 1
	w.Destroy(e1)
	e2 := w.Spawn() // should be ID 2, not 1

	if e2 == e1 {
		t.Errorf("ID was reused: got %d again", e2)
	}
	if e2 != 2 {
		t.Errorf("expected ID 2, got %d", e2)
	}
}
