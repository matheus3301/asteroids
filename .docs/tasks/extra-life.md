# Extra Life Every 10,000 Points

Award one extra life each time the player's score crosses a 10,000-point threshold
(10,000 / 20,000 / 30,000 ...), matching the 1979 Asteroids cabinet behavior.

## Changes

### 1. New field on `Game` struct (`internal/game/game.go`)

```go
type Game struct {
    // ... existing fields ...
    lives           int
    nextExtraLifeAt int   // score threshold for next 1-up
    // ...
}
```

### 2. Reset the threshold in `reset()`

In `reset()`, after `g.lives = 3`, add:

```go
g.nextExtraLifeAt = 10_000
```

### 3. Check after score increments in `updatePlaying()`

Immediately after the `switch ast.Size` block that awards points (still inside the bullet-hit loop), add:

```go
for g.score >= g.nextExtraLifeAt {
    g.lives++
    g.nextExtraLifeAt += 10_000
}
```

Using a `for` loop (not `if`) handles the edge case where a single hit vaults the
score past multiple thresholds -- unlikely with max 100 pts per hit, but correct
by construction.

### 4. No HUD / component / ECS changes required

`g.lives` is already rendered by `drawHUD`. No new components or systems needed.

## Test plan (`internal/game/game_test.go`)

| Test | Setup | Assert |
|------|-------|--------|
| `TestExtraLife_AwardedAt10000` | `g.score = 9_980; g.lives = 3; g.nextExtraLifeAt = 10_000` -- destroy a large asteroid (+20) | `g.lives == 4`, `g.nextExtraLifeAt == 20_000` |
| `TestExtraLife_NotAwardedBelow` | `g.score = 9_900; g.lives = 3; g.nextExtraLifeAt = 10_000` -- destroy a large asteroid (+20) | `g.lives == 3`, `g.nextExtraLifeAt == 10_000` |
| `TestExtraLife_MultipleThresholds` | `g.score = 29_950; g.lives = 3; g.nextExtraLifeAt = 10_000` -- guards the `for` loop | `g.lives == 6`, `g.nextExtraLifeAt == 40_000` |
| `TestReset_ExtraLifeThreshold` | Call `g.reset()` | `g.nextExtraLifeAt == 10_000` |
