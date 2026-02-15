# Bullet Limit (Max 4 on Screen)

The original 1979 Asteroids capped the player at 4 bullets on screen simultaneously. Currently there's no limit, so the player can spam Space.

## Current State

In `internal/game/game.go` at line ~122, every `Space` keypress unconditionally calls `SpawnBullet(w, g.player)`. No bullet count check exists.

`w.bullets` (`internal/game/ecs.go`) is a `map[Entity]*BulletTag` holding all active bullets. Currently all bullets are player bullets -- no UFO bullets exist in the codebase.

## Changes

### 1. Add constant (`internal/game/factory.go`)

```go
const (
    // ... existing constants ...
    MaxPlayerBullets = 4
)
```

Exported so tests can reference it.

### 2. Add count method (`internal/game/ecs.go`)

```go
func (w *World) BulletCount() int {
    return len(w.bullets)
}
```

Centralizes the count. When UFOs are added later, this can be updated to filter by owner.

### 3. Guard the spawn (`internal/game/game.go`)

Change the shooting logic in `updatePlaying()`:

```go
// Handle player shooting
if pc, ok := w.players[g.player]; ok && pc.ShootPressed {
    if w.BulletCount() < MaxPlayerBullets {
        SpawnBullet(w, g.player)
    }
}
```

The guard lives in the game rules layer, not inside `SpawnBullet` (factory layer).

## Future-proofing

When UFOs are added, `BulletTag` should gain an `Owner Entity` field, and `BulletCount` should filter by owner so UFO bullets don't count against the player's limit.

## Test plan (`internal/game/game_test.go`)

| Test | Setup | Assert |
|------|-------|--------|
| `TestBulletLimit_AllowsUpToMax` | Spawn 4 bullets manually | `w.BulletCount() == 4`, 5th spawn is allowed by factory (guard is in game loop) |
| `TestBulletLimit_BlocksWhenAtMax` | Set up game with 4 active bullets, set `pc.ShootPressed = true` | No new bullet spawned, count stays at 4 |
| `TestBulletLimit_AllowsAfterExpiry` | 4 bullets, destroy one, then shoot | New bullet spawned, count back to 4 |
| `TestBulletLimit_AllowsAfterHit` | 4 bullets, destroy one via collision | New bullet can be spawned |
