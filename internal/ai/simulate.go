package ai

import (
	"math/rand"

	"github.com/matheus3301/asteroids/internal/game"
)

// SimConfig holds parameters for a headless simulation run.
type SimConfig struct {
	MaxTicks       int
	NumRuns        int     // average over N runs to reduce noise (default 3)
	ScoreWeight    float64 // multiplier for game score
	SurvivalBonus  float64 // per-tick reward for staying alive
	ApproachWeight float64 // reward for closing distance to asteroids
	HitBonus       float64 // instant bonus per asteroid destroyed
	ShotPenalty    float64 // penalty per wasted shot (fired but nothing in sights)
}

// DefaultSimConfig returns sensible defaults.
func DefaultSimConfig() SimConfig {
	return SimConfig{
		MaxTicks:       3600, // 60 seconds at 60 TPS
		NumRuns:        3,
		ScoreWeight:    3.0,
		SurvivalBonus:  0.05,
		ApproachWeight: 0.2,
		HitBonus:       20.0,
		ShotPenalty:    3.0,
	}
}

// SimResult holds the outcome of a simulation run.
type SimResult struct {
	Score   int
	Ticks   int
	Fitness float64
}

// RunSimulation runs NumRuns headless games and returns the averaged result.
func RunSimulation(agent game.AIAgent, cfg SimConfig) SimResult {
	runs := cfg.NumRuns
	if runs < 1 {
		runs = 1
	}

	totalFitness := 0.0
	totalScore := 0
	totalTicks := 0

	for range runs {
		r := runOnce(agent, cfg)
		totalFitness += r.Fitness
		totalScore += r.Score
		totalTicks += r.Ticks
	}

	return SimResult{
		Score:   totalScore / runs,
		Ticks:   totalTicks / runs,
		Fitness: totalFitness / float64(runs),
	}
}

// runOnce executes a single headless game and returns the result.
func runOnce(agent game.AIAgent, cfg SimConfig) SimResult {
	w := game.NewWorld()
	w.Score = 0
	w.Lives = 3
	w.NextExtraLifeAt = 10_000
	w.Level = 1
	w.SaucerSpawnTimer = 600
	w.SaucerActive = 0
	w.Player = game.SpawnPlayer(w, game.ScreenWidth/2, game.ScreenHeight/2)
	game.SpawnInitialWave(w)

	shapingReward := 0.0
	prevScore := 0

	for tick := range cfg.MaxTicks {
		if w.Lives <= 0 {
			return SimResult{
				Score:   w.Score,
				Ticks:   tick,
				Fitness: computeFitness(w.Score, tick, shapingReward, cfg),
			}
		}

		// AI decides action
		obs := game.ExtractObservation(w)
		action := agent.Act(obs)

		// Track bullets before ShootingSystem to detect fired shots
		bulletsBefore := w.BulletCount()

		// Apply AI input (replaces InputSystem)
		game.AIInputSystem(w, action)

		// Run all game systems except InputSystem and SoundSystem
		game.PhysicsSystem(w)
		game.WrapSystem(w)
		game.InvulnerabilitySystem(w)
		game.LifetimeSystem(w)
		game.SaucerSpawnSystem(w)
		game.SaucerAISystem(w)
		game.SaucerBulletLifetimeSystem(w)
		game.SaucerDespawnSystem(w)
		game.HyperspaceSystem(w, rand.Float64())
		game.ShootingSystem(w)
		events := game.CollisionSystem(w)
		game.CollisionResponseSystem(w, events)
		game.WaveClearSystem(w)

		// --- Fitness shaping ---

		// 1) Approach reward: bonus for moving toward nearest asteroid.
		closingSpeed := obs[35] // normalized by maxSpeed
		if closingSpeed > 0 {
			shapingReward += closingSpeed * cfg.ApproachWeight
		}

		// 2) Shot penalty: penalize shots fired when NOT aimed.
		//    obs[36] is lead line-of-fire quality [0,1]. If it's low and a
		//    bullet was actually spawned, that's a wasted shot.
		bulletsAfter := w.BulletCount()
		shotFired := bulletsAfter > bulletsBefore
		if shotFired {
			lineOfFire := obs[36]
			if lineOfFire < 0.3 {
				shapingReward -= cfg.ShotPenalty
			} else {
				shapingReward += 2.0
			}
		}

		// 3) Big bonus when score actually increases (asteroid destroyed).
		if w.Score > prevScore {
			shapingReward += cfg.HitBonus
		}
		prevScore = w.Score

		// 4) Collision threat penalty: when something is on a collision
		//    course (obs[38] > 0) and the agent isn't thrusting to dodge,
		//    penalize. This teaches "move out of the way."
		threat := obs[38]
		if threat > 0.3 && !action.Thrust {
			shapingReward -= threat * 1.5
		}

		// Clear sound queue (no sound in headless mode)
		w.SoundQueue = w.SoundQueue[:0]
	}

	return SimResult{
		Score:   w.Score,
		Ticks:   cfg.MaxTicks,
		Fitness: computeFitness(w.Score, cfg.MaxTicks, shapingReward, cfg),
	}
}

func computeFitness(score, ticks int, shapingReward float64, cfg SimConfig) float64 {
	return float64(score)*cfg.ScoreWeight +
		float64(ticks)*cfg.SurvivalBonus +
		shapingReward
}
