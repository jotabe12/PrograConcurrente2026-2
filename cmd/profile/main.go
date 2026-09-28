package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	gocode "tp-concurrente/go-code"
)

// profile genera un perfil de CPU de una sola función de predicción y lo escribe
// en el archivo indicado. El caller es responsable de limpiar el archivo.
func profile(path string, fn func()) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create profile file: %w", err)
	}
	defer f.Close()
	if err := pprof.StartCPUProfile(f); err != nil {
		return fmt.Errorf("start CPU profile: %w", err)
	}
	fn()
	pprof.StopCPUProfile()
	return nil
}

func main() {
	csvPath := flag.String("csv", "processed/yellow_tripdata_2026-01.csv", "path to the processed CSV dataset")
	maxRows := flag.Int("max-rows", 100000, "maximum rows to load; use 0 for the complete dataset")
	trainRows := flag.Int("train-rows", 10000, "number of initial rows used for training")
	epochs := flag.Int("epochs", 1, "training epochs")
	learningRate := flag.Float64("learning-rate", 0.001, "training learning rate")
	workers := flag.Int("workers", 8, "number of inference workers for the concurrent profile")
	repeats := flag.Int("repeats", 20, "number of prediction repetitions inside each profile window (aumenta la señal de muestreo)")
	seqOut := flag.String("seq-out", "results/cpu_sequential.prof", "output path for sequential CPU profile")
	poolOut := flag.String("pool-out", "results/cpu_worker_pool.prof", "output path for worker-pool CPU profile")
	flag.Parse()

	// ── Carga y entrenamiento ────────────────────────────────────────────────
	samples, err := gocode.LoadSamplesFromCSV(*csvPath, *maxRows)
	if err != nil {
		log.Fatal(err)
	}
	if *trainRows <= 0 || *trainRows >= len(samples) {
		log.Fatalf("train-rows must be between 1 and %d", len(samples)-1)
	}

	standardizer, err := gocode.FitStandardizer(samples[:*trainRows])
	if err != nil {
		log.Fatal(err)
	}
	normalized := standardizer.Transform(samples)
	network := gocode.NewNeuralNetwork(42)
	if _, err := network.Train(normalized[:*trainRows], *epochs, *learningRate); err != nil {
		log.Fatal(err)
	}
	eval := normalized[*trainRows:]

	// Calentamiento: evita que el JIT y el GC sesquen el primer perfil.
	gocode.PredictSequential(network, eval)
	if _, err := gocode.PredictWorkerPool(network, eval, *workers); err != nil {
		log.Fatal(err)
	}

	// ── Perfil SECUENCIAL ────────────────────────────────────────────────────
	if err := profile(*seqOut, func() {
		for i := 0; i < *repeats; i++ {
			gocode.PredictSequential(network, eval)
		}
	}); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("perfil CPU secuencial  → %s\n", *seqOut)

	// ── Perfil WORKER POOL ───────────────────────────────────────────────────
	if err := profile(*poolOut, func() {
		for i := 0; i < *repeats; i++ {
			if _, err := gocode.PredictWorkerPool(network, eval, *workers); err != nil {
				log.Fatal(err)
			}
		}
	}); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("perfil CPU worker pool → %s\n", *poolOut)

	fmt.Printf("\nPara inspeccionar en la terminal:\n")
	fmt.Printf("  go tool pprof -top %s\n", *seqOut)
	fmt.Printf("  go tool pprof -top %s\n", *poolOut)
	fmt.Printf("\nPara abrir el visualizador interactivo (requiere Graphviz):\n")
	fmt.Printf("  go tool pprof -http=:6060 %s\n", *poolOut)
}
