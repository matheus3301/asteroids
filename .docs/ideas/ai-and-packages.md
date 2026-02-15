# Future: AI Agents & Reusable Go Packages

## Vision

Use Asteroids as a testbed to build reusable Go packages for ECS, neural networks, and genetic algorithms — things the Go ecosystem lacks compared to Python. The game becomes both a product and a showcase.

## Packages to Extract

### `pkg/ecs` — Generic Entity Component System

Extract the ECS pattern already proven in `internal/game/`. Make it generic with Go generics:

- `World`, `Entity`, `Store[T]`, `Get[T]`, `Query[T]`
- Small, opinionated, Go-idiomatic — not trying to be Unity ECS
- Asteroids becomes the first consumer / use case
- Refactor `internal/game/ecs.go` onto it once the API stabilizes

### `pkg/neural` — Feedforward Neural Networks

Pure Go, no cgo, no GPU. Small nets for neuroevolution (e.g. 12→16→5):

- `New(layers ...int)`, `Forward(input) output`
- `Weights() []float64` / `SetWeights([]float64)` for GA integration
- `Save(io.Writer)` / `Load(io.Reader)` with gob encoding
- Activation functions: ReLU, tanh, sigmoid

### `pkg/genetic` — Genetic Algorithm Framework

Generic over genome type using Go generics:

- Population management, tournament selection, crossover, mutation
- `Evolve(evaluate func(G) float64) G` — returns best genome
- Parallelizable evaluation via goroutines
- Configurable: population size, mutation rate, elite count

### `pkg/agent` — Game AI Interface

Bridge between AI and game environments:

- `Agent` interface: `Act(Observation) Action`
- `NeuralAgent` wrapping a `neural.Network`
- Observation/action as `[]float64`

## Current Approach

For now, everything lives under `internal/` to iterate fast:

```
internal/
    game/          # existing game logic
    ai/
        neural.go  # feedforward net
        genetic.go # GA
        agent.go   # agent interface + observation extraction
        simulate.go # headless game loop
```

Once the APIs stabilize, extract to `pkg/` or separate repos.

## Training Pipeline

```
cmd/train/main.go
    1. Create GA population of neural network weight vectors
    2. For each individual:
       - Build Network, set weights
       - Run headless game simulation (no rendering, no sound, max speed)
       - Extract observations from World each tick
       - Agent produces actions, applied to PlayerControl
       - Return fitness = score + survival bonus
    3. Evolve (select, crossover, mutate)
    4. Repeat for N generations
    5. Save best weights to file (gob binary)
```

Training is embarrassingly parallel — each game sim is independent.

## Watch Mode

```
cmd/watch/main.go (or menu option "WATCH AI" in main game)
    1. Load trained weights file
    2. Create NeuralAgent from weights
    3. Run normal game with full rendering and sound
    4. Replace InputSystem: extract observations → Agent.Act() → apply to PlayerControl
```

Could also embed the best weights via `go:embed` so the binary ships with a trained AI.

## Observation Space

What the AI sees each tick:

- Player: position (x, y), velocity (vx, vy), angle
- N nearest asteroids: relative dx, dy, dvx, dvy, size
- Saucer: relative dx, dy (if active)
- Bullet count, lives

## Action Space

5 binary outputs: rotate left, rotate right, thrust, fire, hyperspace.

## Why Go Instead of Python

| Need | Python | Go |
|---|---|---|
| Small neural nets | PyTorch (massive overkill) | Nothing good yet |
| Genetic algorithms | DEAP | Nothing maintained |
| Game AI interface | OpenAI Gym | Nothing |
| Simple ECS | N/A | Scattered attempts |

For networks with <1000 weights, pure Go matrix math is fast enough. No need for GPU, cgo, or Python interop. Training runs can saturate all CPU cores with goroutines trivially.
