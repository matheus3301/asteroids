# Hyperspace (Teleport)

Classic Asteroids panic button. Press Shift to teleport to a random screen position with a small chance (~1/16) of destroying the ship.

## Behaviour

- Press Left Shift or Right Shift to activate
- Player teleports to a random position on screen
- Velocity is zeroed (no momentum carried through)
- 12 departure particles spawn at the old position
- ~6.25% chance of instant death on re-entry (risk/reward)
- 30-tick cooldown (~0.5s) prevents spam
- Allowed while invulnerable (it's a panic button)
- Blocked during cooldown

## Changes

### 1. Component fields (`internal/game/components.go`)

Add to `PlayerControl`:

```go
type PlayerControl struct {
    // ... existing fields ...
    HyperspacePressed  bool
    HyperspaceCooldown int  // ticks remaining before next use
}
```

### 2. Input reading (`internal/game/systems.go`)

In `InputSystem`, after the existing `pc.ShootPressed` line:

```go
pc.HyperspacePressed = inpututil.IsKeyJustPressed(ebiten.KeyShiftLeft) ||
    inpututil.IsKeyJustPressed(ebiten.KeyShiftRight)
```

### 3. Cooldown tick (`internal/game/systems.go`)

In `InvulnerabilitySystem`, outside the invulnerability check (runs every tick):

```go
if pc.HyperspaceCooldown > 0 {
    pc.HyperspaceCooldown--
}
```

### 4. Core logic (`internal/game/game.go`)

In `updatePlaying()`, after `InputSystem(w)` and before `PhysicsSystem(w)`:

```go
// Hyperspace
if pc, ok := w.players[g.player]; ok && pc.HyperspacePressed && pc.HyperspaceCooldown <= 0 {
    pos := w.positions[g.player]
    vel := w.velocities[g.player]

    // Departure particles
    for i := 0; i < 12; i++ {
        SpawnParticle(w, pos.X, pos.Y)
    }

    // Risk: ~1/16 chance of death
    if rand.Float64() < 1.0/16.0 {
        // Same death logic as player-asteroid collision
        g.lives--
        if g.lives <= 0 {
            g.state = stateGameOver
            w.Destroy(g.player)
        } else {
            // Respawn at center with invulnerability
            pos.X = ScreenWidth / 2
            pos.Y = ScreenHeight / 2
            vel.X, vel.Y = 0, 0
            // ... set invulnerability (same as existing respawn code)
        }
    } else {
        // Successful teleport
        pos.X = rand.Float64() * ScreenWidth
        pos.Y = rand.Float64() * ScreenHeight
        vel.X, vel.Y = 0, 0
    }

    pc.HyperspaceCooldown = 30
}
```

### Files changed

| File | Change |
|------|--------|
| `internal/game/components.go` | Add `HyperspacePressed`, `HyperspaceCooldown` to `PlayerControl` |
| `internal/game/systems.go` | Read Shift key in `InputSystem`, tick cooldown in `InvulnerabilitySystem` |
| `internal/game/game.go` | Hyperspace logic block in `updatePlaying()` |
| `internal/game/systems_test.go` | New tests |

No new files needed.

## Edge cases

| Scenario | Behaviour |
|----------|-----------|
| Press Shift during cooldown | Nothing happens |
| Press Shift while invulnerable | Teleport works (it's a panic button) |
| Land inside an asteroid | Normal collision resolves next frame |
| Hyperspace kills player at 1 life | Game over |
| Hyperspace kills player at 2+ lives | Respawn at center with invulnerability |

## Test plan (`internal/game/systems_test.go` and `game_test.go`)

| Test | Assert |
|------|--------|
| `TestHyperspace_TeleportsToNewPosition` | Position changes after hyperspace |
| `TestHyperspace_ZerosVelocity` | vel.X == 0, vel.Y == 0 after teleport |
| `TestHyperspace_SetsCooldown` | `pc.HyperspaceCooldown == 30` after use |
| `TestHyperspace_BlockedDuringCooldown` | Position unchanged when cooldown > 0 |
| `TestHyperspace_CooldownDecrement` | Cooldown decreases each tick via InvulnerabilitySystem |
| `TestHyperspace_AllowedWhileInvulnerable` | Teleport works with `pc.Invulnerable = true` |
| `TestHyperspace_DeathOnBadLuck` | Seed RNG for death outcome, verify lives decremented |
| `TestHyperspace_SpawnsDepartureParticles` | Particle count increases after teleport |
