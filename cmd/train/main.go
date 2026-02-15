package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/matheus3301/asteroids/internal/ai"
	"github.com/matheus3301/asteroids/internal/game"
)

func main() {
	generations := flag.Int("generations", 200, "number of generations to train")
	population := flag.Int("population", 300, "population size")
	output := flag.String("output", "weights.gob", "output file for best weights")
	workers := flag.Int("workers", 0, "number of parallel workers (0 = NumCPU)")
	maxTicks := flag.Int("max-ticks", 3600, "max ticks per simulation (60 tps)")
	numRuns := flag.Int("runs", 3, "evaluations per genome (averaged to reduce noise)")
	hiddenStr := flag.String("hidden", "16", "hidden layer sizes, comma-separated (e.g. 16 or 20,10)")
	flag.Parse()

	hidden := parseHidden(*hiddenStr)
	layers := make([]int, 0, 2+len(hidden))
	layers = append(layers, game.ObservationSize)
	layers = append(layers, hidden...)
	layers = append(layers, 5)

	net := ai.NewNetwork(layers...)
	fmt.Printf("Network: %s (%d parameters)\n", formatLayers(layers), net.WeightCount())

	cfg := ai.DefaultGAConfig(net.WeightCount())
	cfg.PopulationSize = *population
	if *workers > 0 {
		cfg.Workers = *workers
	}

	simCfg := ai.DefaultSimConfig()
	simCfg.MaxTicks = *maxTicks
	simCfg.NumRuns = *numRuns

	fmt.Printf("Population: %d, Generations: %d, Workers: %d, Runs/eval: %d\n",
		cfg.PopulationSize, *generations, cfg.Workers, simCfg.NumRuns)

	ga := ai.NewGA(cfg)

	for gen := range *generations {
		start := time.Now()

		ga.Evolve(func(genome []float64) float64 {
			n := ai.NewNetwork(layers...)
			n.SetParams(genome)
			agent := ai.NewNeuralAgent(n)
			result := ai.RunSimulation(agent, simCfg)
			return result.Fitness
		})

		elapsed := time.Since(start)
		fmt.Printf("Gen %3d | Best: %8.1f | BestEver: %8.1f | Avg: %8.1f | Time: %.1fs\n",
			gen+1,
			ga.Population[0].Fitness,
			ga.BestEver.Fitness,
			avgFitness(ga.Population),
			elapsed.Seconds(),
		)
	}

	// Save best network
	bestNet := ai.NewNetwork(layers...)
	bestNet.SetParams(ga.BestEver.Genome)

	f, err := os.Create(*output)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
	}
	defer f.Close()

	if err := bestNet.Save(f); err != nil {
		log.Fatalf("failed to save weights: %v", err)
	}

	fmt.Printf("\nBest fitness: %.1f\n", ga.BestEver.Fitness)
	fmt.Printf("Weights saved to %s\n", *output)
}

func avgFitness(pop []ai.Individual) float64 {
	sum := 0.0
	for _, ind := range pop {
		sum += ind.Fitness
	}
	return sum / float64(len(pop))
}

func parseHidden(s string) []int {
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			log.Fatalf("invalid hidden layer size %q: %v", p, err)
		}
		result = append(result, n)
	}
	return result
}

func formatLayers(layers []int) string {
	parts := make([]string, len(layers))
	for i, l := range layers {
		parts[i] = strconv.Itoa(l)
	}
	return strings.Join(parts, " -> ")
}
