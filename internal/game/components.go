package game

import "image/color"

// Position represents a location in 2D space.
type Position struct {
	X, Y float64
}

// Velocity represents movement in 2D space.
type Velocity struct {
	X, Y float64
}

// Rotation represents an angle and optional spin rate.
type Rotation struct {
	Angle float64
	Spin  float64
}

// Collider represents a circular collision boundary.
type Collider struct {
	Radius float64
}

// ShapeKind identifies how to render an entity.
type ShapeKind int

const (
	ShapeTriangle ShapeKind = iota
	ShapePolygon
	ShapeCircle
)

// Renderable holds drawing data for an entity.
type Renderable struct {
	Kind     ShapeKind
	Vertices [][2]float64 // local-space vertices (for polygon/triangle)
	Color    color.RGBA
	Scale    float64
}

// PlayerControl marks an entity as player-controlled and holds player state.
type PlayerControl struct {
	Thrusting         bool
	ShootPressed      bool
	Invulnerable      bool
	InvulnerableTimer int
	BlinkTimer        int
}

// AsteroidSize represents the three asteroid sizes.
type AsteroidSize int

const (
	SizeLarge  AsteroidSize = 0
	SizeMedium AsteroidSize = 1
	SizeSmall  AsteroidSize = 2
)

// AsteroidTag marks an entity as an asteroid.
type AsteroidTag struct {
	Size AsteroidSize
}

// BulletTag marks an entity as a bullet with a lifetime.
type BulletTag struct {
	Life int
}

// ParticleTag marks an entity as a particle with a lifetime.
type ParticleTag struct {
	Life    int
	MaxLife int
}

// SaucerSize represents the two saucer variants.
type SaucerSize int

const (
	SaucerLarge SaucerSize = iota
	SaucerSmall
)

// SaucerTag marks an entity as a flying saucer.
type SaucerTag struct {
	Size          SaucerSize
	DirectionX    float64 // +1.0 (moving right) or -1.0 (moving left)
	ShootCooldown int     // ticks remaining until next shot
	VerticalTimer int     // ticks until the next vertical direction change
}

// SaucerBulletTag marks an entity as a saucer-fired bullet.
type SaucerBulletTag struct {
	Life int
}
