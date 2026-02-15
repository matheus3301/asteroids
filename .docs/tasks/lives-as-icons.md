# Lives Displayed as Ship Icons

Replace the textual "LIVES: N" HUD element with small ship silhouettes, matching the 1979 Asteroids arcade original.

## Current Behaviour

`drawHUD` in `game.go` renders three text lines at the top-left:

```
SCORE: 0      (10, 10)
LIVES: 3      (10, 28)
LEVEL: 1      (10, 46)
```

## Design

### Ship icon vertices

Reuse the same proportions from `SpawnPlayer` (`factory.go`), scaled down:

```go
const hudIconScale = 0.45 // 15 * 0.45 ~ 7px nose-to-tail

var shipIconVerts = [][2]float64{
    {playerRadius * hudIconScale, 0},
    {-playerRadius * 0.8 * hudIconScale, -playerRadius * 0.6 * hudIconScale},
    {-playerRadius * 0.8 * hudIconScale, playerRadius * 0.6 * hudIconScale},
}
```

Produces a triangle ~13px wide by ~8px tall, similar to text glyph height at `hudScale=2.0` (14px).

### How many icons

Draw `g.lives - 1` icons. The ship on screen IS the current life; only reserve lives are shown. When `g.lives <= 1`, draw zero icons.

### Positioning

- Same Y row where "LIVES: 3" was (`y ~ 35`, vertically centered in the row)
- Start X at `10` (same left margin)
- Space each icon `20px` apart (13px width + 7px gap)
- Point upward (angle = `-math.Pi/2`) to match default ship orientation

### Drawing

Call existing `drawPolygon` from `render.go`:

```go
func (g *Game) drawHUD(screen *ebiten.Image) {
    hudScale := 2.0
    hudColor := color.RGBA{255, 255, 255, 255}

    DrawText(screen, fmt.Sprintf("SCORE: %d", g.score), 10, 10, hudScale, hudColor)

    // Lives as ship icons
    iconY := 35.0
    for i := 0; i < g.lives-1; i++ {
        iconX := 10.0 + float64(i)*20.0
        pos := &Position{X: iconX, Y: iconY}
        drawPolygon(screen, pos, -math.Pi/2, shipIconVerts, hudColor)
    }

    DrawText(screen, fmt.Sprintf("LEVEL: %d", g.level), 10, 46, hudScale, hudColor)
}
```

No new rendering helpers needed.

### Files changed

| File | Change |
|------|--------|
| `internal/game/game.go` | Update `drawHUD`: replace LIVES text with icon loop, add `shipIconVerts`/`hudIconScale` |

## Test plan

| Test | Assert |
|------|--------|
| `TestLivesIconCount_3Lives` | `g.lives = 3` -> 2 icons drawn |
| `TestLivesIconCount_1Life` | `g.lives = 1` -> 0 icons |
| `TestLivesIconCount_0Lives` | `g.lives = 0` -> 0 icons (game over edge) |
| `TestShipIconVerts_HasThreeVertices` | `len(shipIconVerts) == 3` |
| Visual smoke test | Run game, verify icons appear/disappear correctly |
