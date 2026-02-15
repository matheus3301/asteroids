# UFO / Flying Saucer Enemy -- Implementation Spec

## Overview

Add the classic UFO (flying saucer) enemy from the 1979 Asteroids arcade game.
Two variants exist: a **Large saucer** (inaccurate fire, 200 pts) and a **Small
saucer** (aims directly at the player, 1000 pts). Only one UFO may be active at
any time. The UFO spawns from a screen edge, travels horizontally with periodic
vertical direction changes, fires bullets, and is destroyed by a single player
bullet. UFO bullets and body contact kill the player.

Reference behaviour from the original arcade game:
- Large saucer fires in a random direction.
- Small saucer leads its shots toward the player.
- At low scores the large saucer appears; as the player's score increases the
  small saucer becomes increasingly likely.
- The saucer enters from the left or right edge and exits from the opposite
  side. If it reaches the far edge without being destroyed, it despawns.

---

## 1. New Components (`components.go`)

### 1.1 `SaucerSize` type

```go
type SaucerSize int

const (
    SaucerLarge SaucerSize = iota
    SaucerSmall
)
```

### 1.2 `SaucerTag` component

```go
type SaucerTag struct {
    Size           SaucerSize
    DirectionX     float64 // +1.0 (moving right) or -1.0 (moving left)
    ShootCooldown  int     // ticks remaining until next shot
    VerticalTimer  int     // ticks until the next vertical direction change
}
```

### 1.3 `SaucerBulletTag` component

UFO bullets need to be distinguished from player bullets so collision logic
can tell them apart. The existing `BulletTag` is used only for player bullets.

```go
type SaucerBulletTag struct {
    Life int
}
```

---

## 2. ECS Changes (`ecs.go`)

### 2.1 New component stores in `World`

```go
saucers       map[Entity]*SaucerTag
saucerBullets map[Entity]*SaucerBulletTag
```

### 2.2 `NewWorld()` -- initialise the new maps

```go
saucers:       make(map[Entity]*SaucerTag),
saucerBullets: make(map[Entity]*SaucerBulletTag),
```

### 2.3 `Destroy()` -- clean up the new maps

```go
delete(w.saucers, e)
delete(w.saucerBullets, e)
```

---

## 3. Constants (`factory.go`)

```go
const (
    saucerLargeRadius      = 20.0
    saucerSmallRadius      = 10.0
    saucerLargeSpeed       = 1.5
    saucerSmallSpeed       = 2.5
    saucerShootCooldownMin = 60   // ~1s
    saucerShootCooldownMax = 150  // ~2.5s
    saucerBulletSpeed      = 4.0
    saucerBulletLife       = 90   // ~1.5s
    saucerVerticalTimerMin = 60
    saucerVerticalTimerMax = 180
    saucerVerticalSpeed    = 0.8
)
```

---

## 4. Factory Functions (`factory.go`)

### 4.1 `SpawnSaucer(w *World, size SaucerSize) Entity`

- Enter from left or right edge (random), just off-screen
- Random Y in the middle 60% of the screen
- Radius and speed based on size
- Saucer polygon via `saucerVertices(radius)` helper
- Color: red (`{255,0,0,255}`)
- NOT added to `wrappers` (despawns at opposite edge)
- `SaucerTag` with size, direction, shoot cooldown, vertical timer

### 4.2 `SpawnSaucerBullet(w *World, saucerEntity Entity, px, py float64) Entity`

- Large saucer: random angle
- Small saucer: aimed at player position `(px, py)`
- Spawns at saucer position, velocity at `saucerBulletSpeed`
- `ShapeCircle`, reddish color (`{255,100,100,255}`)
- Added to `wrappers`
- `SaucerBulletTag{Life: saucerBulletLife}`

---

## 5. Systems (`systems.go`)

### 5.1 New: `SaucerAISystem(w *World, playerPos *Position)`

Per saucer each tick:
1. **Shoot cooldown**: Decrement, fire bullet at 0, reset to random value
2. **Vertical direction**: Decrement timer, pick new Y velocity from `{-0.8, 0, +0.8}` when expired
3. **Vertical wrap**: If Y < 0 or Y > ScreenHeight, wrap
4. **Despawn at far edge**: If past opposite screen edge + radius, destroy

### 5.2 New: `SaucerBulletLifetimeSystem(w *World)`

Same pattern as bullet lifetime: decrement Life, destroy at 0.

### 5.3 Extend: `CollisionSystem`

New event types:
```go
type saucerHit struct { Bullet, Saucer Entity }

type CollisionEvent struct {
    // ... existing ...
    SaucerBulletHits []saucerHit  // player bullets hitting saucers
}
```

New collision checks:
1. **Player Bullet vs Saucer**: point-in-circle
2. **Saucer Bullet vs Player**: sets `PlayerHit = true`
3. **Saucer Body vs Player**: circle-circle, sets `PlayerHit = true`

---

## 6. Game Loop (`game.go`)

### New fields

```go
saucerSpawnTimer int    // ticks until next saucer spawn
saucerActive     Entity // current saucer (0 = none)
```

### Constants

```go
const (
    saucerInitialDelay = 600  // ~10s before first saucer
    saucerRespawnDelay = 600  // ~10s between saucers
)
```

### `reset()`: Init `saucerSpawnTimer = saucerInitialDelay`, `saucerActive = 0`

### `updatePlaying()` additions:
1. Spawn timer: if no active saucer, decrement timer, spawn at 0
2. `chooseSaucerSize(score)`: large below 10K, small above 40K, linear mix between
3. Call `SaucerAISystem` and `SaucerBulletLifetimeSystem`
4. Process `events.SaucerBulletHits`: score (200/1000), 12 particles, destroy
5. On player death: destroy saucer + all saucer bullets, reset timer

### `Draw()`: Add `DrawSaucerDetail(g.world, screen)` alongside `DrawThrust`

---

## 7. Rendering (`render.go`)

### `saucerVertices(radius float64) [][2]float64`

Classic saucer outline: bottom point, lower curves, wide rim, dome, back around.

### `DrawSaucerDetail(w *World, screen *ebiten.Image)`

For each saucer, draw two horizontal interior lines:
- Rim line (full width at Y=0)
- Dome base line (narrower, above rim)

Follows the `DrawThrust` pattern.

---

## 8. Edge Cases

1. Only one saucer at a time (enforced by `saucerActive`)
2. Saucer does NOT wrap horizontally -- despawns at opposite edge
3. Saucer bullets DO wrap (added to `wrappers`)
4. Saucer survives wave transitions
5. Saucer + bullets destroyed on player death (clean respawn)
6. Initial 10s delay before first saucer
7. Pausing freezes saucer (updatePlaying not called)
8. Saucer does not collide with asteroids
9. Entity ID 0 is never valid (safe sentinel)

---

## 9. Test Plan

### Factory tests
- SpawnSaucer: all components, edge spawn, radius per size, velocity direction, not a wrapper, polygon renderable
- SpawnSaucerBullet: all components, aimed vs random, lifetime, is wrapper, speed magnitude

### System tests
- SaucerAISystem: cooldown decrement, fires at 0, cooldown resets, vertical timer, direction change, vertical wrap, despawn at edge, nil player safe
- SaucerBulletLifetimeSystem: decrement, destroy at 0
- CollisionSystem: bullet hits/misses saucer, saucer bullet hits/misses player, invulnerable skip, body collision, expired bullet ignored

### Game tests
- Spawn timer decrements, triggers at 0, no spawn while active
- chooseSaucerSize: low/mid/high score
- Destruction: score increment, saucerActive cleared, particles
- Player death: saucer cleared, bullets cleared, timer reset
- Wave clear: saucer survives
- reset(): clears saucer state

### ECS tests
- Destroy cleans saucer maps
- NewWorld initializes saucer maps

---

## 10. Implementation Order

1. Components (`components.go`)
2. ECS maps (`ecs.go`)
3. Saucer vertices (`render.go`)
4. Factory functions + constants (`factory.go`)
5. AI system + bullet lifetime (`systems.go`)
6. Collision system extension (`systems.go`)
7. Saucer detail rendering (`render.go`)
8. Game loop integration (`game.go`)
9. All tests
10. Playtest and tune constants

## Files Changed

| File | Change |
|------|--------|
| `internal/game/components.go` | Add `SaucerSize`, `SaucerTag`, `SaucerBulletTag` |
| `internal/game/ecs.go` | Add maps, init, destroy |
| `internal/game/factory.go` | Constants, `SpawnSaucer`, `SpawnSaucerBullet` |
| `internal/game/systems.go` | `SaucerAISystem`, `SaucerBulletLifetimeSystem`, extend `CollisionSystem` |
| `internal/game/render.go` | `saucerVertices`, `DrawSaucerDetail` |
| `internal/game/game.go` | Fields, `reset()`, spawn timer, AI calls, event handling, `chooseSaucerSize` |
