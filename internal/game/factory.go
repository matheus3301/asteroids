package game

import (
	"image/color"
	"math"
	"math/rand"
)

const (
	playerRadius     = 15.0
	bulletSpeed      = 7.0
	bulletLife       = 60
	MaxPlayerBullets = 4

	saucerLargeRadius      = 20.0
	saucerSmallRadius      = 10.0
	saucerLargeSpeed       = 1.5
	saucerSmallSpeed       = 2.5
	saucerShootCooldownMin = 60
	saucerShootCooldownMax = 150
	saucerBulletSpeed      = 4.0
	saucerBulletLife       = 90
	saucerVerticalTimerMin = 60
	saucerVerticalTimerMax = 180
	saucerVerticalSpeed    = 0.8
)

// SpawnPlayer creates the player ship entity.
func SpawnPlayer(w *World, x, y float64) Entity {
	e := w.Spawn()

	w.positions[e] = &Position{X: x, Y: y}
	w.velocities[e] = &Velocity{}
	w.rotations[e] = &Rotation{Angle: -math.Pi / 2}
	w.colliders[e] = &Collider{Radius: playerRadius}
	w.wrappers[e] = true

	// Triangle vertices (local space, pointing right at angle=0)
	w.renderables[e] = &Renderable{
		Kind: ShapeTriangle,
		Vertices: [][2]float64{
			{playerRadius, 0},                          // nose
			{-playerRadius * 0.8, -playerRadius * 0.6}, // left
			{-playerRadius * 0.8, playerRadius * 0.6},  // right
		},
		Color: color.RGBA{0, 255, 0, 255},
		Scale: 1,
	}

	w.players[e] = &PlayerControl{
		Invulnerable:      true,
		InvulnerableTimer: 120,
	}

	return e
}

// SpawnAsteroid creates an asteroid entity.
func SpawnAsteroid(w *World, x, y float64, size AsteroidSize) Entity {
	e := w.Spawn()

	var radius, speed float64
	switch size {
	case SizeLarge:
		radius = 40
		speed = 1.0
	case SizeMedium:
		radius = 20
		speed = 1.8
	case SizeSmall:
		radius = 10
		speed = 2.5
	}

	dir := rand.Float64() * 2 * math.Pi
	spd := speed * (0.5 + rand.Float64())

	w.positions[e] = &Position{X: x, Y: y}
	w.velocities[e] = &Velocity{
		X: math.Cos(dir) * spd,
		Y: math.Sin(dir) * spd,
	}
	w.rotations[e] = &Rotation{
		Spin: (rand.Float64() - 0.5) * 0.04,
	}
	w.colliders[e] = &Collider{Radius: radius}
	w.wrappers[e] = true

	// Generate irregular polygon vertices
	numVerts := 8 + rand.Intn(5)
	verts := make([][2]float64, numVerts)
	for i := range verts {
		ang := float64(i) / float64(numVerts) * 2 * math.Pi
		r := radius * (0.7 + rand.Float64()*0.3)
		verts[i] = [2]float64{math.Cos(ang) * r, math.Sin(ang) * r}
	}

	w.renderables[e] = &Renderable{
		Kind:     ShapePolygon,
		Vertices: verts,
		Color:    color.RGBA{200, 200, 200, 255},
		Scale:    1,
	}

	w.asteroids[e] = &AsteroidTag{Size: size}

	return e
}

// SpawnBullet creates a bullet fired from the player.
func SpawnBullet(w *World, playerEntity Entity) Entity {
	e := w.Spawn()

	pos := w.positions[playerEntity]
	rot := w.rotations[playerEntity]

	w.positions[e] = &Position{
		X: pos.X + math.Cos(rot.Angle)*playerRadius,
		Y: pos.Y + math.Sin(rot.Angle)*playerRadius,
	}
	w.velocities[e] = &Velocity{
		X: math.Cos(rot.Angle) * bulletSpeed,
		Y: math.Sin(rot.Angle) * bulletSpeed,
	}
	w.colliders[e] = &Collider{Radius: 2}
	w.wrappers[e] = true

	w.renderables[e] = &Renderable{
		Kind:  ShapeCircle,
		Color: color.RGBA{255, 255, 255, 255},
		Scale: 2,
	}

	w.bullets[e] = &BulletTag{Life: bulletLife}

	return e
}

// SpawnSaucer creates a flying saucer that enters from a screen edge.
func SpawnSaucer(w *World, size SaucerSize) Entity {
	e := w.Spawn()

	var radius, speed float64
	switch size {
	case SaucerLarge:
		radius = saucerLargeRadius
		speed = saucerLargeSpeed
	case SaucerSmall:
		radius = saucerSmallRadius
		speed = saucerSmallSpeed
	}

	// Enter from left or right edge
	dirX := 1.0
	x := -radius
	if rand.Intn(2) == 0 {
		dirX = -1.0
		x = ScreenWidth + radius
	}
	// Random Y in middle 60% of screen
	y := ScreenHeight*0.2 + rand.Float64()*ScreenHeight*0.6

	w.positions[e] = &Position{X: x, Y: y}
	w.velocities[e] = &Velocity{X: dirX * speed, Y: 0}
	w.rotations[e] = &Rotation{}
	w.colliders[e] = &Collider{Radius: radius}

	verts := saucerVertices(radius)
	w.renderables[e] = &Renderable{
		Kind:     ShapePolygon,
		Vertices: verts,
		Color:    color.RGBA{255, 0, 0, 255},
		Scale:    1,
	}

	w.saucers[e] = &SaucerTag{
		Size:          size,
		DirectionX:    dirX,
		ShootCooldown: saucerShootCooldownMin + rand.Intn(saucerShootCooldownMax-saucerShootCooldownMin),
		VerticalTimer: saucerVerticalTimerMin + rand.Intn(saucerVerticalTimerMax-saucerVerticalTimerMin),
	}

	return e
}

// SpawnSaucerBullet creates a bullet fired by a saucer. px, py is the player position (for aimed shots).
func SpawnSaucerBullet(w *World, saucerEntity Entity, px, py float64) Entity {
	e := w.Spawn()

	spos := w.positions[saucerEntity]
	st := w.saucers[saucerEntity]

	var angle float64
	if st.Size == SaucerSmall {
		// Aim at player
		dx := px - spos.X
		dy := py - spos.Y
		angle = math.Atan2(dy, dx)
	} else {
		// Random direction
		angle = rand.Float64() * 2 * math.Pi
	}

	w.positions[e] = &Position{X: spos.X, Y: spos.Y}
	w.velocities[e] = &Velocity{
		X: math.Cos(angle) * saucerBulletSpeed,
		Y: math.Sin(angle) * saucerBulletSpeed,
	}
	w.colliders[e] = &Collider{Radius: 2}
	w.wrappers[e] = true

	w.renderables[e] = &Renderable{
		Kind:  ShapeCircle,
		Color: color.RGBA{255, 100, 100, 255},
		Scale: 2,
	}

	w.saucerBullets[e] = &SaucerBulletTag{Life: saucerBulletLife}

	return e
}

// SpawnParticle creates an explosion particle.
func SpawnParticle(w *World, x, y float64) Entity {
	e := w.Spawn()

	angle := rand.Float64() * 2 * math.Pi
	speed := 1.0 + rand.Float64()*3.0
	life := 20 + rand.Intn(20)

	w.positions[e] = &Position{X: x, Y: y}
	w.velocities[e] = &Velocity{
		X: math.Cos(angle) * speed,
		Y: math.Sin(angle) * speed,
	}

	w.renderables[e] = &Renderable{
		Kind:  ShapeCircle,
		Color: color.RGBA{255, 200, 50, 255},
		Scale: 1.5,
	}

	w.particles[e] = &ParticleTag{Life: life, MaxLife: life}

	return e
}
