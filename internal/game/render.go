package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func strokeLine(screen *ebiten.Image, x1, y1, x2, y2 float64, clr color.Color) {
	vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 1.5, clr, false)
}

// RenderSystem draws all renderable entities.
func RenderSystem(w *World, screen *ebiten.Image) {
	for e, r := range w.renderables {
		pos := w.positions[e]
		if pos == nil {
			continue
		}

		// Check blink for invulnerable players
		if pc, ok := w.players[e]; ok {
			if pc.Invulnerable && (pc.BlinkTimer/8)%2 == 0 {
				continue
			}
		}

		// Particle alpha fade
		clr := r.Color
		if pt, ok := w.particles[e]; ok {
			alpha := float64(pt.Life) / float64(pt.MaxLife) * 255
			if alpha < 0 {
				alpha = 0
			}
			clr.A = uint8(alpha)
		}

		rot := w.rotations[e]
		angle := 0.0
		if rot != nil {
			angle = rot.Angle
		}

		switch r.Kind {
		case ShapeTriangle, ShapePolygon:
			drawPolygon(screen, pos, angle, r.Vertices, clr)
		case ShapeCircle:
			vector.FillCircle(screen, float32(pos.X), float32(pos.Y), float32(r.Scale), clr, false)
		}
	}
}

func drawPolygon(screen *ebiten.Image, pos *Position, angle float64, verts [][2]float64, clr color.RGBA) {
	n := len(verts)
	if n < 2 {
		return
	}
	cos := math.Cos(angle)
	sin := math.Sin(angle)

	transform := func(v [2]float64) (float64, float64) {
		return pos.X + v[0]*cos - v[1]*sin,
			pos.Y + v[0]*sin + v[1]*cos
	}

	for i := 0; i < n; i++ {
		x1, y1 := transform(verts[i])
		x2, y2 := transform(verts[(i+1)%n])
		strokeLine(screen, x1, y1, x2, y2, clr)
	}
}

// saucerVertices generates a classic flying saucer outline polygon.
func saucerVertices(radius float64) [][2]float64 {
	r := radius
	return [][2]float64{
		{-r * 0.4, r * 0.5},  // bottom-left
		{r * 0.4, r * 0.5},   // bottom-right
		{r * 0.8, r * 0.1},   // lower-right rim
		{r, -r * 0.1},        // right rim tip
		{r * 0.6, -r * 0.3},  // upper-right rim
		{r * 0.3, -r * 0.7},  // dome right
		{0, -r},               // dome top
		{-r * 0.3, -r * 0.7}, // dome left
		{-r * 0.6, -r * 0.3}, // upper-left rim
		{-r, -r * 0.1},       // left rim tip
		{-r * 0.8, r * 0.1},  // lower-left rim
	}
}

// DrawSaucerDetail draws interior detail lines on saucers (rim + dome base).
func DrawSaucerDetail(w *World, screen *ebiten.Image) {
	for e := range w.saucers {
		pos := w.positions[e]
		r := w.renderables[e]
		if pos == nil || r == nil {
			continue
		}

		radius := w.colliders[e].Radius
		clr := r.Color

		// Rim line (full width at Y=0)
		strokeLine(screen, pos.X-radius, pos.Y-radius*0.1, pos.X+radius, pos.Y-radius*0.1, clr)
		// Dome base line (narrower, above rim)
		strokeLine(screen, pos.X-radius*0.6, pos.Y-radius*0.3, pos.X+radius*0.6, pos.Y-radius*0.3, clr)
	}
}

// DrawThrust draws the flame behind the player ship.
func DrawThrust(w *World, screen *ebiten.Image) {
	for e, pc := range w.players {
		if !pc.Thrusting {
			continue
		}
		if pc.Invulnerable && (pc.BlinkTimer/8)%2 == 0 {
			continue
		}
		pos := w.positions[e]
		rot := w.rotations[e]
		r := w.renderables[e]
		if pos == nil || rot == nil || r == nil || len(r.Vertices) < 3 {
			continue
		}

		cos := math.Cos(rot.Angle)
		sin := math.Sin(rot.Angle)
		transform := func(v [2]float64) (float64, float64) {
			return pos.X + v[0]*cos - v[1]*sin,
				pos.Y + v[0]*sin + v[1]*cos
		}

		lx, ly := transform(r.Vertices[1])
		rx, ry := transform(r.Vertices[2])

		tailX := pos.X - cos*playerRadius*1.2
		tailY := pos.Y - sin*playerRadius*1.2

		flameClr := color.RGBA{255, 165, 0, 255}
		strokeLine(screen, lx, ly, tailX, tailY, flameClr)
		strokeLine(screen, rx, ry, tailX, tailY, flameClr)
	}
}
