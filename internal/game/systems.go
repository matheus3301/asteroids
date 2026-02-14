package game

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	rotationSpeed = 0.05
	thrustPower   = 0.12
	maxSpeed      = 5.0
	friction      = 0.99
	particleDrag  = 0.96
)

// InputSystem reads keyboard input and updates player entities.
func InputSystem(w *World) {
	for e, pc := range w.players {
		rot := w.rotations[e]
		vel := w.velocities[e]

		if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
			rot.Angle -= rotationSpeed
		}
		if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
			rot.Angle += rotationSpeed
		}

		pc.Thrusting = ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW)
		if pc.Thrusting {
			vel.X += math.Cos(rot.Angle) * thrustPower
			vel.Y += math.Sin(rot.Angle) * thrustPower
			speed := math.Sqrt(vel.X*vel.X + vel.Y*vel.Y)
			if speed > maxSpeed {
				vel.X = vel.X / speed * maxSpeed
				vel.Y = vel.Y / speed * maxSpeed
			}
		}

		// Friction on player
		vel.X *= friction
		vel.Y *= friction

		pc.ShootPressed = inpututil.IsKeyJustPressed(ebiten.KeySpace)
		pc.HyperspacePressed = inpututil.IsKeyJustPressed(ebiten.KeyShiftLeft) ||
			inpututil.IsKeyJustPressed(ebiten.KeyShiftRight)
	}
}

// PhysicsSystem applies velocity to position and spin to rotation.
func PhysicsSystem(w *World) {
	for e, pos := range w.positions {
		if vel, ok := w.velocities[e]; ok {
			pos.X += vel.X
			pos.Y += vel.Y
		}
		if rot, ok := w.rotations[e]; ok {
			rot.Angle += rot.Spin
		}
	}
}

// WrapSystem wraps entities around the screen edges.
func WrapSystem(w *World) {
	for e := range w.wrappers {
		pos := w.positions[e]
		if pos == nil {
			continue
		}
		if pos.X < 0 {
			pos.X += ScreenWidth
		} else if pos.X > ScreenWidth {
			pos.X -= ScreenWidth
		}
		if pos.Y < 0 {
			pos.Y += ScreenHeight
		} else if pos.Y > ScreenHeight {
			pos.Y -= ScreenHeight
		}
	}
}

// LifetimeSystem decrements bullet and particle lifetimes and destroys expired ones.
func LifetimeSystem(w *World) {
	for e, b := range w.bullets {
		b.Life--
		if b.Life <= 0 {
			w.Destroy(e)
		}
	}
	for e, p := range w.particles {
		p.Life--
		if p.Life <= 0 {
			w.Destroy(e)
		}
		// Apply drag to particles
		if vel, ok := w.velocities[e]; ok {
			vel.X *= particleDrag
			vel.Y *= particleDrag
		}
	}
}

// InvulnerabilitySystem ticks down player invulnerability timers.
func InvulnerabilitySystem(w *World) {
	for _, pc := range w.players {
		if pc.Invulnerable {
			pc.BlinkTimer++
			pc.InvulnerableTimer--
			if pc.InvulnerableTimer <= 0 {
				pc.Invulnerable = false
			}
		}
	}
}

// SaucerAISystem updates saucer behavior: shooting, vertical movement, despawn.
func SaucerAISystem(w *World, playerPos *Position) {
	for e, st := range w.saucers {
		pos := w.positions[e]
		vel := w.velocities[e]
		if pos == nil || vel == nil {
			continue
		}

		// Shoot cooldown
		st.ShootCooldown--
		if st.ShootCooldown <= 0 {
			px, py := 0.0, 0.0
			if playerPos != nil {
				px, py = playerPos.X, playerPos.Y
			}
			SpawnSaucerBullet(w, e, px, py)
			st.ShootCooldown = saucerShootCooldownMin + rand.Intn(saucerShootCooldownMax-saucerShootCooldownMin)
		}

		// Vertical direction changes
		st.VerticalTimer--
		if st.VerticalTimer <= 0 {
			choices := []float64{-saucerVerticalSpeed, 0, saucerVerticalSpeed}
			vel.Y = choices[rand.Intn(3)]
			st.VerticalTimer = saucerVerticalTimerMin + rand.Intn(saucerVerticalTimerMax-saucerVerticalTimerMin)
		}

		// Vertical wrap
		if pos.Y < 0 {
			pos.Y += ScreenHeight
		} else if pos.Y > ScreenHeight {
			pos.Y -= ScreenHeight
		}

		// Despawn at far edge
		col := w.colliders[e]
		radius := 0.0
		if col != nil {
			radius = col.Radius
		}
		if st.DirectionX > 0 && pos.X > ScreenWidth+radius {
			w.Destroy(e)
		} else if st.DirectionX < 0 && pos.X < -radius {
			w.Destroy(e)
		}
	}
}

// SaucerBulletLifetimeSystem decrements saucer bullet lifetimes and destroys expired ones.
func SaucerBulletLifetimeSystem(w *World) {
	for e, sb := range w.saucerBullets {
		sb.Life--
		if sb.Life <= 0 {
			w.Destroy(e)
		}
	}
}

// CollisionEvent describes what happened during a collision check.
type CollisionEvent struct {
	BulletHits       []bulletHit
	SaucerBulletHits []saucerHit
	PlayerHit        bool
	PlayerEntity     Entity
}

type bulletHit struct {
	Bullet   Entity
	Asteroid Entity
}

type saucerHit struct {
	Bullet Entity
	Saucer Entity
}

// CollisionSystem checks bullet-asteroid and player-asteroid collisions.
func CollisionSystem(w *World) CollisionEvent {
	var events CollisionEvent

	// Bullet vs Asteroid
	for be, bt := range w.bullets {
		if bt.Life <= 0 {
			continue
		}
		bpos := w.positions[be]
		if bpos == nil {
			continue
		}
		for ae := range w.asteroids {
			apos := w.positions[ae]
			acol := w.colliders[ae]
			if apos == nil || acol == nil {
				continue
			}
			dx := bpos.X - apos.X
			dy := bpos.Y - apos.Y
			if dx*dx+dy*dy < acol.Radius*acol.Radius {
				events.BulletHits = append(events.BulletHits, bulletHit{
					Bullet:   be,
					Asteroid: ae,
				})
				break
			}
		}
	}

	// Player Bullet vs Saucer
	for be, bt := range w.bullets {
		if bt.Life <= 0 {
			continue
		}
		bpos := w.positions[be]
		if bpos == nil {
			continue
		}
		for se := range w.saucers {
			spos := w.positions[se]
			scol := w.colliders[se]
			if spos == nil || scol == nil {
				continue
			}
			dx := bpos.X - spos.X
			dy := bpos.Y - spos.Y
			if dx*dx+dy*dy < scol.Radius*scol.Radius {
				events.SaucerBulletHits = append(events.SaucerBulletHits, saucerHit{
					Bullet: be,
					Saucer: se,
				})
				break
			}
		}
	}

	// Player vs Asteroid
	for pe, pc := range w.players {
		if pc.Invulnerable {
			continue
		}
		ppos := w.positions[pe]
		pcol := w.colliders[pe]
		if ppos == nil || pcol == nil {
			continue
		}
		for ae := range w.asteroids {
			apos := w.positions[ae]
			acol := w.colliders[ae]
			if apos == nil || acol == nil {
				continue
			}
			dx := ppos.X - apos.X
			dy := ppos.Y - apos.Y
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < pcol.Radius+acol.Radius {
				events.PlayerHit = true
				events.PlayerEntity = pe
				return events
			}
		}

		// Saucer Bullet vs Player
		for sbe := range w.saucerBullets {
			sbpos := w.positions[sbe]
			if sbpos == nil {
				continue
			}
			dx := ppos.X - sbpos.X
			dy := ppos.Y - sbpos.Y
			if dx*dx+dy*dy < pcol.Radius*pcol.Radius {
				events.PlayerHit = true
				events.PlayerEntity = pe
				return events
			}
		}

		// Saucer Body vs Player
		for se := range w.saucers {
			spos := w.positions[se]
			scol := w.colliders[se]
			if spos == nil || scol == nil {
				continue
			}
			dx := ppos.X - spos.X
			dy := ppos.Y - spos.Y
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < pcol.Radius+scol.Radius {
				events.PlayerHit = true
				events.PlayerEntity = pe
				return events
			}
		}
	}

	return events
}
