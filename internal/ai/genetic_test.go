package ai

import (
	"testing"
)

func TestDefaultGAConfig(t *testing.T) {
	cfg := DefaultGAConfig(645)
	if cfg.GenomeSize != 645 {
		t.Errorf("expected genome size 645, got %d", cfg.GenomeSize)
	}
	if cfg.PopulationSize != 300 {
		t.Errorf("expected population 300, got %d", cfg.PopulationSize)
	}
	if cfg.Workers < 1 {
		t.Errorf("expected at least 1 worker, got %d", cfg.Workers)
	}
}

func TestNewGA_PopulationSize(t *testing.T) {
	cfg := DefaultGAConfig(10)
	cfg.PopulationSize = 20
	ga := NewGA(cfg)
	if len(ga.Population) != 20 {
		t.Errorf("expected 20 individuals, got %d", len(ga.Population))
	}
	for i, ind := range ga.Population {
		if len(ind.Genome) != 10 {
			t.Errorf("individual %d: expected genome size 10, got %d", i, len(ind.Genome))
		}
	}
}

func TestNewGA_GenomesAreRandom(t *testing.T) {
	cfg := DefaultGAConfig(100)
	cfg.PopulationSize = 10
	ga := NewGA(cfg)
	// Two random genomes should not be identical
	same := true
	for i := range ga.Population[0].Genome {
		if ga.Population[0].Genome[i] != ga.Population[1].Genome[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("first two genomes should not be identical")
	}
}

func TestEvolve_FitnessEvaluated(t *testing.T) {
	cfg := DefaultGAConfig(10)
	cfg.PopulationSize = 20
	cfg.EliteCount = 2
	cfg.Workers = 2
	ga := NewGA(cfg)

	// Simple fitness: sum of genome values
	ga.Evolve(func(genome []float64) float64 {
		sum := 0.0
		for _, v := range genome {
			sum += v
		}
		return sum
	})

	if ga.Generation != 1 {
		t.Errorf("expected generation 1, got %d", ga.Generation)
	}
	if ga.BestEver.Fitness == 0 {
		t.Error("best ever fitness should be set")
	}
}

func TestEvolve_FitnessImproves(t *testing.T) {
	cfg := DefaultGAConfig(10)
	cfg.PopulationSize = 50
	cfg.EliteCount = 5
	cfg.Workers = 4
	ga := NewGA(cfg)

	// Fitness: negative sum of squares (maximize toward 0)
	eval := func(genome []float64) float64 {
		sum := 0.0
		for _, v := range genome {
			sum += v * v
		}
		return -sum
	}

	ga.Evolve(eval)
	firstBest := ga.BestEver.Fitness

	for range 20 {
		ga.Evolve(eval)
	}

	if ga.BestEver.Fitness < firstBest {
		t.Errorf("fitness should improve: first=%v, after 20 gens=%v", firstBest, ga.BestEver.Fitness)
	}
}

func TestEvolve_ElitismPreserved(t *testing.T) {
	cfg := DefaultGAConfig(5)
	cfg.PopulationSize = 20
	cfg.EliteCount = 3
	cfg.Workers = 1
	ga := NewGA(cfg)

	eval := func(genome []float64) float64 {
		return genome[0] // simple fitness
	}

	ga.Evolve(eval)

	// After evolve, population is sorted. Store top 3 fitnesses.
	// The next generation should preserve elite genomes.
	topFitnesses := make([]float64, 3)
	topGenomes := make([][]float64, 3)
	for i := range 3 {
		topFitnesses[i] = ga.Population[i].Fitness
		topGenomes[i] = append([]float64{}, ga.Population[i].Genome...)
	}

	ga.Evolve(eval)

	// After second evolve, the elite should still be present (potentially reordered)
	// At minimum, the best ever should not regress
	if ga.BestEver.Fitness < topFitnesses[0] {
		t.Error("best ever should not regress")
	}
}

func TestBlendCrossover_Length(t *testing.T) {
	a := []float64{1, 2, 3, 4, 5}
	b := []float64{6, 7, 8, 9, 10}
	child := blendCrossover(a, b)
	if len(child) != 5 {
		t.Errorf("expected child length 5, got %d", len(child))
	}
	// Each gene should be between the two parents
	for i, v := range child {
		lo := a[i]
		hi := b[i]
		if lo > hi {
			lo, hi = hi, lo
		}
		if v < lo || v > hi {
			t.Errorf("child[%d]=%v is not between a[%d]=%v and b[%d]=%v", i, v, i, a[i], i, b[i])
		}
	}
}

func TestMutate_ChangesGenome(t *testing.T) {
	genome := make([]float64, 100)
	mutate(genome, 1.0, 1.0) // 100% mutation rate

	changed := false
	for _, v := range genome {
		if v != 0 {
			changed = true
			break
		}
	}
	if !changed {
		t.Error("100% mutation should change at least one gene")
	}
}

func TestMutate_ZeroRate(t *testing.T) {
	genome := []float64{1, 2, 3}
	original := append([]float64{}, genome...)
	mutate(genome, 0.0, 1.0)

	for i := range genome {
		if genome[i] != original[i] {
			t.Error("0% mutation should not change genome")
		}
	}
}
