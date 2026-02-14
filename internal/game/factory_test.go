package game

import (
	"image/color"
	"math"
	"testing"
)

// --------------- SpawnPlayer ---------------

func TestSpawnPlayer_HasAllComponents(t *testing.T) {
	w := NewWorld()
	e := SpawnPlayer(w, 400, 300)

	if w.positions[e] == nil {
		t.Error("missing position component")
	}
	if w.velocities[e] == nil {
		t.Error("missing velocity component")
	}
	if w.rotations[e] == nil {
		t.Error("missing rotation component")
	}
	if w.colliders[e] == nil {
		t.Error("missing collider component")
	}
	if w.renderables[e] == nil {
		t.Error("missing renderable component")
	}
	if w.players[e] == nil {
		t.Error("missing player control component")
	}
	if !w.wrappers[e] {
		t.Error("player should be a wrapper entity")
	}
}

func TestSpawnPlayer_InitialPosition(t *testing.T) {
	w := NewWorld()
	e := SpawnPlayer(w, 400, 300)

	pos := w.positions[e]
	if pos.X != 400 || pos.Y != 300 {
		t.Errorf("expected position (400,300), got (%v,%v)", pos.X, pos.Y)
	}
}

func TestSpawnPlayer_InitialAngle(t *testing.T) {
	w := NewWorld()
	e := SpawnPlayer(w, 400, 300)

	rot := w.rotations[e]
	expected := -math.Pi / 2
	if math.Abs(rot.Angle-expected) > 1e-9 {
		t.Errorf("expected angle %v, got %v", expected, rot.Angle)
	}
}

func TestSpawnPlayer_StartsInvulnerable(t *testing.T) {
	w := NewWorld()
	e := SpawnPlayer(w, 400, 300)

	pc := w.players[e]
	if !pc.Invulnerable {
		t.Error("player should start invulnerable")
	}
	if pc.InvulnerableTimer != 120 {
		t.Errorf("expected invulnerability timer 120, got %d", pc.InvulnerableTimer)
	}
}

func TestSpawnPlayer_ColliderRadius(t *testing.T) {
	w := NewWorld()
	e := SpawnPlayer(w, 400, 300)

	col := w.colliders[e]
	if col.Radius != 15 {
		t.Errorf("expected collider radius 15, got %v", col.Radius)
	}
}

func TestSpawnPlayer_ZeroInitialVelocity(t *testing.T) {
	w := NewWorld()
	e := SpawnPlayer(w, 400, 300)

	vel := w.velocities[e]
	if vel.X != 0 || vel.Y != 0 {
		t.Errorf("expected zero velocity, got (%v,%v)", vel.X, vel.Y)
	}
}

func TestSpawnPlayer_RenderableIsGreenTriangle(t *testing.T) {
	w := NewWorld()
	e := SpawnPlayer(w, 400, 300)

	r := w.renderables[e]
	if r.Kind != ShapeTriangle {
		t.Errorf("expected ShapeTriangle, got %v", r.Kind)
	}
	if r.Color != (color.RGBA{0, 255, 0, 255}) {
		t.Errorf("expected green, got %v", r.Color)
	}
	if len(r.Vertices) != 3 {
		t.Errorf("expected 3 vertices, got %d", len(r.Vertices))
	}
}

// --------------- SpawnAsteroid ---------------

func TestSpawnAsteroid_SizeLarge(t *testing.T) {
	w := NewWorld()
	e := SpawnAsteroid(w, 200, 200, SizeLarge)

	col := w.colliders[e]
	if col.Radius != 40 {
		t.Errorf("large asteroid: expected radius 40, got %v", col.Radius)
	}
	ast := w.asteroids[e]
	if ast.Size != SizeLarge {
		t.Errorf("expected SizeLarge, got %v", ast.Size)
	}
}

func TestSpawnAsteroid_SizeMedium(t *testing.T) {
	w := NewWorld()
	e := SpawnAsteroid(w, 200, 200, SizeMedium)

	col := w.colliders[e]
	if col.Radius != 20 {
		t.Errorf("medium asteroid: expected radius 20, got %v", col.Radius)
	}
	ast := w.asteroids[e]
	if ast.Size != SizeMedium {
		t.Errorf("expected SizeMedium, got %v", ast.Size)
	}
}

func TestSpawnAsteroid_SizeSmall(t *testing.T) {
	w := NewWorld()
	e := SpawnAsteroid(w, 200, 200, SizeSmall)

	col := w.colliders[e]
	if col.Radius != 10 {
		t.Errorf("small asteroid: expected radius 10, got %v", col.Radius)
	}
	ast := w.asteroids[e]
	if ast.Size != SizeSmall {
		t.Errorf("expected SizeSmall, got %v", ast.Size)
	}
}

func TestSpawnAsteroid_HasAllComponents(t *testing.T) {
	w := NewWorld()
	e := SpawnAsteroid(w, 100, 100, SizeLarge)

	if w.positions[e] == nil {
		t.Error("missing position")
	}
	if w.velocities[e] == nil {
		t.Error("missing velocity")
	}
	if w.rotations[e] == nil {
		t.Error("missing rotation")
	}
	if w.colliders[e] == nil {
		t.Error("missing collider")
	}
	if w.renderables[e] == nil {
		t.Error("missing renderable")
	}
	if w.asteroids[e] == nil {
		t.Error("missing asteroid tag")
	}
	if !w.wrappers[e] {
		t.Error("asteroid should be a wrapper entity")
	}
}

func TestSpawnAsteroid_HasNonZeroVelocity(t *testing.T) {
	w := NewWorld()
	e := SpawnAsteroid(w, 100, 100, SizeLarge)

	vel := w.velocities[e]
	speed := math.Sqrt(vel.X*vel.X + vel.Y*vel.Y)
	if speed == 0 {
		t.Error("asteroid should have non-zero velocity")
	}
}

func TestSpawnAsteroid_PolygonVertices(t *testing.T) {
	w := NewWorld()
	e := SpawnAsteroid(w, 100, 100, SizeLarge)

	r := w.renderables[e]
	if r.Kind != ShapePolygon {
		t.Errorf("expected ShapePolygon, got %v", r.Kind)
	}
	if len(r.Vertices) < 8 || len(r.Vertices) > 12 {
		t.Errorf("expected 8-12 vertices, got %d", len(r.Vertices))
	}
}

func TestSpawnAsteroid_SpeedScalesWithSize(t *testing.T) {
	// Run multiple times to check that small asteroids are generally faster
	var largeSpeeds, smallSpeeds []float64
	for i := 0; i < 20; i++ {
		w := NewWorld()
		el := SpawnAsteroid(w, 100, 100, SizeLarge)
		vl := w.velocities[el]
		largeSpeeds = append(largeSpeeds, math.Sqrt(vl.X*vl.X+vl.Y*vl.Y))

		es := SpawnAsteroid(w, 100, 100, SizeSmall)
		vs := w.velocities[es]
		smallSpeeds = append(smallSpeeds, math.Sqrt(vs.X*vs.X+vs.Y*vs.Y))
	}

	var avgLarge, avgSmall float64
	for i := range largeSpeeds {
		avgLarge += largeSpeeds[i]
		avgSmall += smallSpeeds[i]
	}
	avgLarge /= float64(len(largeSpeeds))
	avgSmall /= float64(len(smallSpeeds))

	if avgSmall <= avgLarge {
		t.Errorf("small asteroids should be faster on average: large=%v, small=%v", avgLarge, avgSmall)
	}
}

// --------------- SpawnBullet ---------------

func TestSpawnBullet_SpawnsAtPlayerNose(t *testing.T) {
	w := NewWorld()
	player := SpawnPlayer(w, 400, 300)
	// Player faces up (-π/2)

	bullet := SpawnBullet(w, player)

	bpos := w.positions[bullet]
	expectedX := 400 + math.Cos(-math.Pi/2)*15
	expectedY := 300 + math.Sin(-math.Pi/2)*15

	if math.Abs(bpos.X-expectedX) > 1e-9 || math.Abs(bpos.Y-expectedY) > 1e-9 {
		t.Errorf("expected bullet at (%v,%v), got (%v,%v)", expectedX, expectedY, bpos.X, bpos.Y)
	}
}

func TestSpawnBullet_VelocityInPlayerDirection(t *testing.T) {
	w := NewWorld()
	player := SpawnPlayer(w, 400, 300)
	// Player faces up (-π/2)

	bullet := SpawnBullet(w, player)

	vel := w.velocities[bullet]
	expectedVX := math.Cos(-math.Pi/2) * 7.0
	expectedVY := math.Sin(-math.Pi/2) * 7.0

	if math.Abs(vel.X-expectedVX) > 1e-9 || math.Abs(vel.Y-expectedVY) > 1e-9 {
		t.Errorf("expected bullet velocity (%v,%v), got (%v,%v)", expectedVX, expectedVY, vel.X, vel.Y)
	}
}

func TestSpawnBullet_HasCorrectLifetime(t *testing.T) {
	w := NewWorld()
	player := SpawnPlayer(w, 400, 300)
	bullet := SpawnBullet(w, player)

	bt := w.bullets[bullet]
	if bt.Life != 60 {
		t.Errorf("expected bullet life 60, got %d", bt.Life)
	}
}

func TestSpawnBullet_HasCollider(t *testing.T) {
	w := NewWorld()
	player := SpawnPlayer(w, 400, 300)
	bullet := SpawnBullet(w, player)

	col := w.colliders[bullet]
	if col == nil {
		t.Fatal("bullet should have a collider")
	}
	if col.Radius != 2 {
		t.Errorf("expected bullet collider radius 2, got %v", col.Radius)
	}
}

func TestSpawnBullet_HasRenderable(t *testing.T) {
	w := NewWorld()
	player := SpawnPlayer(w, 400, 300)
	bullet := SpawnBullet(w, player)

	r := w.renderables[bullet]
	if r == nil {
		t.Fatal("bullet should have a renderable")
	}
	if r.Kind != ShapeCircle {
		t.Errorf("expected ShapeCircle, got %v", r.Kind)
	}
	if r.Color != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("expected white, got %v", r.Color)
	}
}

func TestSpawnBullet_IsWrapped(t *testing.T) {
	w := NewWorld()
	player := SpawnPlayer(w, 400, 300)
	bullet := SpawnBullet(w, player)

	if !w.wrappers[bullet] {
		t.Error("bullet should be a wrapper entity")
	}
}

// --------------- SpawnSaucer ---------------

func TestSpawnSaucer_HasAllComponents(t *testing.T) {
	w := NewWorld()
	e := SpawnSaucer(w, SaucerLarge)

	if w.positions[e] == nil {
		t.Error("missing position")
	}
	if w.velocities[e] == nil {
		t.Error("missing velocity")
	}
	if w.rotations[e] == nil {
		t.Error("missing rotation")
	}
	if w.colliders[e] == nil {
		t.Error("missing collider")
	}
	if w.renderables[e] == nil {
		t.Error("missing renderable")
	}
	if w.saucers[e] == nil {
		t.Error("missing saucer tag")
	}
}

func TestSpawnSaucer_NotAWrapper(t *testing.T) {
	w := NewWorld()
	e := SpawnSaucer(w, SaucerLarge)

	if w.wrappers[e] {
		t.Error("saucer should NOT be a wrapper entity")
	}
}

func TestSpawnSaucer_LargeRadius(t *testing.T) {
	w := NewWorld()
	e := SpawnSaucer(w, SaucerLarge)

	col := w.colliders[e]
	if col.Radius != 20 {
		t.Errorf("expected large saucer radius 20, got %v", col.Radius)
	}
}

func TestSpawnSaucer_SmallRadius(t *testing.T) {
	w := NewWorld()
	e := SpawnSaucer(w, SaucerSmall)

	col := w.colliders[e]
	if col.Radius != 10 {
		t.Errorf("expected small saucer radius 10, got %v", col.Radius)
	}
}

func TestSpawnSaucer_EdgeSpawn(t *testing.T) {
	// Run multiple times to verify spawns are at edges
	for i := 0; i < 20; i++ {
		w := NewWorld()
		e := SpawnSaucer(w, SaucerLarge)
		pos := w.positions[e]
		col := w.colliders[e]

		atLeft := pos.X == -col.Radius
		atRight := pos.X == ScreenWidth+col.Radius
		if !atLeft && !atRight {
			t.Errorf("saucer should spawn at edge, got X=%v", pos.X)
		}
	}
}

func TestSpawnSaucer_VelocityMatchesDirection(t *testing.T) {
	for i := 0; i < 20; i++ {
		w := NewWorld()
		e := SpawnSaucer(w, SaucerLarge)
		vel := w.velocities[e]
		st := w.saucers[e]

		if st.DirectionX > 0 && vel.X <= 0 {
			t.Error("saucer moving right should have positive X velocity")
		}
		if st.DirectionX < 0 && vel.X >= 0 {
			t.Error("saucer moving left should have negative X velocity")
		}
	}
}

func TestSpawnSaucer_PolygonRenderable(t *testing.T) {
	w := NewWorld()
	e := SpawnSaucer(w, SaucerLarge)

	r := w.renderables[e]
	if r.Kind != ShapePolygon {
		t.Errorf("expected ShapePolygon, got %v", r.Kind)
	}
	if len(r.Vertices) < 8 {
		t.Errorf("expected at least 8 vertices for saucer shape, got %d", len(r.Vertices))
	}
	if r.Color != (color.RGBA{255, 0, 0, 255}) {
		t.Errorf("expected red, got %v", r.Color)
	}
}

// --------------- SpawnSaucerBullet ---------------

func TestSpawnSaucerBullet_HasAllComponents(t *testing.T) {
	w := NewWorld()
	saucer := SpawnSaucer(w, SaucerLarge)
	e := SpawnSaucerBullet(w, saucer, 400, 300)

	if w.positions[e] == nil {
		t.Error("missing position")
	}
	if w.velocities[e] == nil {
		t.Error("missing velocity")
	}
	if w.colliders[e] == nil {
		t.Error("missing collider")
	}
	if w.renderables[e] == nil {
		t.Error("missing renderable")
	}
	if w.saucerBullets[e] == nil {
		t.Error("missing saucer bullet tag")
	}
}

func TestSpawnSaucerBullet_IsWrapper(t *testing.T) {
	w := NewWorld()
	saucer := SpawnSaucer(w, SaucerLarge)
	e := SpawnSaucerBullet(w, saucer, 400, 300)

	if !w.wrappers[e] {
		t.Error("saucer bullet should be a wrapper entity")
	}
}

func TestSpawnSaucerBullet_Lifetime(t *testing.T) {
	w := NewWorld()
	saucer := SpawnSaucer(w, SaucerLarge)
	e := SpawnSaucerBullet(w, saucer, 400, 300)

	sb := w.saucerBullets[e]
	if sb.Life != 90 {
		t.Errorf("expected saucer bullet life 90, got %d", sb.Life)
	}
}

func TestSpawnSaucerBullet_SpeedMagnitude(t *testing.T) {
	w := NewWorld()
	saucer := SpawnSaucer(w, SaucerLarge)
	e := SpawnSaucerBullet(w, saucer, 400, 300)

	vel := w.velocities[e]
	speed := math.Sqrt(vel.X*vel.X + vel.Y*vel.Y)
	if math.Abs(speed-4.0) > 0.01 {
		t.Errorf("expected saucer bullet speed 4.0, got %v", speed)
	}
}

func TestSpawnSaucerBullet_SmallSaucerAimsAtPlayer(t *testing.T) {
	w := NewWorld()
	saucer := SpawnSaucer(w, SaucerSmall)
	// Force saucer position to known location
	w.positions[saucer] = &Position{X: 100, Y: 100}

	e := SpawnSaucerBullet(w, saucer, 200, 100) // player is to the right

	vel := w.velocities[e]
	// Bullet should be going mostly rightward
	if vel.X <= 0 {
		t.Errorf("small saucer bullet aimed at player to the right should have positive X velocity, got %v", vel.X)
	}
	// Y should be near zero since player is at same height
	if math.Abs(vel.Y) > 0.1 {
		t.Errorf("small saucer bullet should have near-zero Y when player at same height, got %v", vel.Y)
	}
}

func TestSpawnSaucerBullet_ReddishColor(t *testing.T) {
	w := NewWorld()
	saucer := SpawnSaucer(w, SaucerLarge)
	e := SpawnSaucerBullet(w, saucer, 400, 300)

	r := w.renderables[e]
	if r.Color != (color.RGBA{255, 100, 100, 255}) {
		t.Errorf("expected reddish color, got %v", r.Color)
	}
}

// --------------- SpawnParticle ---------------

func TestSpawnParticle_HasPosition(t *testing.T) {
	w := NewWorld()
	e := SpawnParticle(w, 100, 200)

	pos := w.positions[e]
	if pos == nil {
		t.Fatal("particle should have a position")
	}
	if pos.X != 100 || pos.Y != 200 {
		t.Errorf("expected position (100,200), got (%v,%v)", pos.X, pos.Y)
	}
}

func TestSpawnParticle_HasVelocity(t *testing.T) {
	w := NewWorld()
	e := SpawnParticle(w, 100, 200)

	vel := w.velocities[e]
	if vel == nil {
		t.Fatal("particle should have a velocity")
	}
	speed := math.Sqrt(vel.X*vel.X + vel.Y*vel.Y)
	if speed < 1.0 || speed > 4.0 {
		t.Errorf("particle speed should be between 1 and 4, got %v", speed)
	}
}

func TestSpawnParticle_HasRenderable(t *testing.T) {
	w := NewWorld()
	e := SpawnParticle(w, 100, 200)

	r := w.renderables[e]
	if r == nil {
		t.Fatal("particle should have a renderable")
	}
	if r.Kind != ShapeCircle {
		t.Errorf("expected ShapeCircle, got %v", r.Kind)
	}
}

func TestSpawnParticle_HasParticleTag(t *testing.T) {
	w := NewWorld()
	e := SpawnParticle(w, 100, 200)

	pt := w.particles[e]
	if pt == nil {
		t.Fatal("particle should have a particle tag")
	}
	if pt.Life < 20 || pt.Life > 39 {
		t.Errorf("particle life should be in [20,39], got %d", pt.Life)
	}
	if pt.Life != pt.MaxLife {
		t.Errorf("initial Life (%d) should equal MaxLife (%d)", pt.Life, pt.MaxLife)
	}
}
