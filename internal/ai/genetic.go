package ai

import (
	"math/rand"
	"runtime"
	"sort"
	"sync"
)

// GAConfig holds parameters for the genetic algorithm.
type GAConfig struct {
	PopulationSize int
	GenomeSize     int
	MutationRate   float64
	MutationScale  float64 // gaussian stddev for mutations
	EliteCount     int
	TournamentSize int
	CrossoverRate  float64
	Workers        int
}

// DefaultGAConfig returns sensible defaults for the given genome size.
func DefaultGAConfig(genomeSize int) GAConfig {
	return GAConfig{
		PopulationSize: 300,
		GenomeSize:     genomeSize,
		MutationRate:   0.10,
		MutationScale:  0.3,
		EliteCount:     20,
		TournamentSize: 5,
		CrossoverRate:  0.3,
		Workers:        runtime.NumCPU(),
	}
}

// Individual represents one member of the population.
type Individual struct {
	Genome  []float64
	Fitness float64
}

// EvalFunc evaluates a genome and returns its fitness.
type EvalFunc func(genome []float64) float64

// GA is the genetic algorithm runner.
type GA struct {
	Config     GAConfig
	Population []Individual
	Generation int
	BestEver   Individual
}

// NewGA creates a new GA with a random initial population.
func NewGA(cfg GAConfig) *GA {
	pop := make([]Individual, cfg.PopulationSize)
	for i := range pop {
		genome := make([]float64, cfg.GenomeSize)
		for j := range genome {
			genome[j] = rand.NormFloat64() * 0.5
		}
		pop[i] = Individual{Genome: genome}
	}
	return &GA{
		Config:     cfg,
		Population: pop,
	}
}

// Evolve runs one generation: evaluate all individuals, select, crossover, mutate.
func (ga *GA) Evolve(evaluate EvalFunc) {
	cfg := ga.Config

	// Parallel evaluation
	var wg sync.WaitGroup
	sem := make(chan struct{}, cfg.Workers)
	for i := range ga.Population {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int) {
			defer wg.Done()
			defer func() { <-sem }()
			ga.Population[idx].Fitness = evaluate(ga.Population[idx].Genome)
		}(i)
	}
	wg.Wait()

	// Sort by fitness descending
	sort.Slice(ga.Population, func(i, j int) bool {
		return ga.Population[i].Fitness > ga.Population[j].Fitness
	})

	// Track best ever
	if ga.Population[0].Fitness > ga.BestEver.Fitness || ga.Generation == 0 {
		ga.BestEver = Individual{
			Genome:  append([]float64{}, ga.Population[0].Genome...),
			Fitness: ga.Population[0].Fitness,
		}
	}

	// Build next generation
	next := make([]Individual, cfg.PopulationSize)

	// Elitism: copy top individuals
	for i := range cfg.EliteCount {
		next[i] = Individual{
			Genome:  append([]float64{}, ga.Population[i].Genome...),
			Fitness: ga.Population[i].Fitness,
		}
	}

	// Fill rest via tournament selection + crossover + mutation
	for i := cfg.EliteCount; i < cfg.PopulationSize; i++ {
		p1 := ga.tournamentSelect()

		var child []float64
		if rand.Float64() < cfg.CrossoverRate {
			p2 := ga.tournamentSelect()
			child = blendCrossover(p1.Genome, p2.Genome)
		} else {
			child = append([]float64{}, p1.Genome...)
		}

		mutate(child, cfg.MutationRate, cfg.MutationScale)
		next[i] = Individual{Genome: child}
	}

	ga.Population = next
	ga.Generation++
}

// tournamentSelect picks the best individual from a random tournament.
func (ga *GA) tournamentSelect() Individual {
	best := ga.Population[rand.Intn(len(ga.Population))]
	for range ga.Config.TournamentSize - 1 {
		contender := ga.Population[rand.Intn(len(ga.Population))]
		if contender.Fitness > best.Fitness {
			best = contender
		}
	}
	return best
}

// blendCrossover averages the two parents with a random blend factor per gene.
// This preserves the general structure of both parents rather than randomly
// swapping individual weights, which destroys learned neural network functions.
func blendCrossover(a, b []float64) []float64 {
	child := make([]float64, len(a))
	for i := range child {
		alpha := rand.Float64() // [0, 1)
		child[i] = alpha*a[i] + (1-alpha)*b[i]
	}
	return child
}

// mutate applies gaussian mutation to each gene with the given probability.
func mutate(genome []float64, rate, scale float64) {
	for i := range genome {
		if rand.Float64() < rate {
			genome[i] += rand.NormFloat64() * scale
		}
	}
}
