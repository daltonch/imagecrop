package main

import (
	"flag"
	"fmt"
	"imagecrop/cropper"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type job struct {
	inputPath string
	filename  string
	outputDir string
	tolerance float64
	maxCrop   float64
}

type result struct {
	filename   string
	success    bool
	wasCropped bool
	message    string
}

func main() {
	// Define CLI flags
	inputDir := flag.String("input", "", "Input directory containing image files (required)")
	outputDir := flag.String("output", "cropped", "Output directory (default: cropped)")
	tolerance := flag.Float64("tolerance", 15.0, "Brightness variation tolerance percentage (0-100, default: 15)")
	maxCrop := flag.Float64("max-crop", 30.0, "Maximum crop percentage per dimension (0-100, default: 30)")
	threads := flag.Int("threads", 4, "Number of concurrent threads (default: 4)")

	flag.Parse()

	// Validate required flags
	if *inputDir == "" {
		fmt.Println("Error: --input flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Validate tolerance
	if *tolerance < 0 || *tolerance > 100 {
		fmt.Println("Error: --tolerance must be between 0 and 100")
		flag.Usage()
		os.Exit(1)
	}

	// Validate max-crop
	if *maxCrop < 0 || *maxCrop > 100 {
		fmt.Println("Error: --max-crop must be between 0 and 100")
		flag.Usage()
		os.Exit(1)
	}

	// Validate threads
	if *threads < 1 {
		fmt.Println("Error: --threads must be at least 1")
		flag.Usage()
		os.Exit(1)
	}

	// Check if input directory exists
	if _, err := os.Stat(*inputDir); os.IsNotExist(err) {
		fmt.Printf("Error: Input directory '%s' does not exist\n", *inputDir)
		os.Exit(1)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Collect all image files first
	var jobs []job
	err := filepath.WalkDir(*inputDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-image files
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			return nil
		}

		jobs = append(jobs, job{
			inputPath: path,
			filename:  filepath.Base(path),
			outputDir: *outputDir,
			tolerance: *tolerance,
			maxCrop:   *maxCrop,
		})

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	if len(jobs) == 0 {
		fmt.Println("\nNo image files found to process.")
		return
	}

	fmt.Printf("Found %d images to process using %d threads...\n\n", len(jobs), *threads)

	// Create channels for jobs and results
	jobChan := make(chan job, len(jobs))
	resultChan := make(chan result, len(jobs))

	// Counters with mutex for thread safety
	var (
		processedCount int
		croppedCount   int
		unchangedCount int
		errorCount     int
		mu             sync.Mutex
		outputMu       sync.Mutex // Separate mutex for console output
	)

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < *threads; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := range jobChan {
				// Print processing message (thread-safe)
				outputMu.Lock()
				fmt.Printf("Processing: %s\n", j.filename)
				outputMu.Unlock()

				// Process the image with a temporary output path
				tempPath := filepath.Join(j.outputDir, fmt.Sprintf(".temp_%d_%s", workerID, j.filename))
				cropResult, err := cropper.CropImage(j.inputPath, tempPath, j.tolerance, j.maxCrop)

				if err != nil {
					outputMu.Lock()
					fmt.Printf("  Error: %v\n", err)
					outputMu.Unlock()

					mu.Lock()
					errorCount++
					mu.Unlock()

					resultChan <- result{
						filename: j.filename,
						success:  false,
						message:  err.Error(),
					}
					continue
				}

				// Determine final output path based on whether image was cropped
				var outputPath string
				if cropResult.WasCropped {
					nameWithoutExt := strings.TrimSuffix(j.filename, filepath.Ext(j.filename))
					outputPath = filepath.Join(j.outputDir, nameWithoutExt+"_cropped"+filepath.Ext(j.filename))
				} else {
					outputPath = filepath.Join(j.outputDir, j.filename)
				}

				// Rename temp file to final output path
				if err := os.Rename(tempPath, outputPath); err != nil {
					outputMu.Lock()
					fmt.Printf("  Error renaming output file: %v\n", err)
					outputMu.Unlock()

					os.Remove(tempPath) // Clean up temp file

					mu.Lock()
					errorCount++
					mu.Unlock()

					resultChan <- result{
						filename: j.filename,
						success:  false,
						message:  err.Error(),
					}
					continue
				}

				// Update counters
				mu.Lock()
				processedCount++
				if cropResult.WasCropped {
					croppedCount++
				} else {
					unchangedCount++
				}
				mu.Unlock()

				// Print result message (thread-safe)
				outputMu.Lock()
				fmt.Printf("  %s -> %s\n", cropResult.Message, filepath.Base(outputPath))
				outputMu.Unlock()

				resultChan <- result{
					filename:   j.filename,
					success:    true,
					wasCropped: cropResult.WasCropped,
					message:    cropResult.Message,
				}
			}
		}(i)
	}

	// Send jobs to workers
	for _, j := range jobs {
		jobChan <- j
	}
	close(jobChan)

	// Wait for all workers to complete
	wg.Wait()
	close(resultChan)

	// Drain results channel (optional, for future use)
	for range resultChan {
		// Results already processed in workers
	}

	// Print summary
	fmt.Printf("\nProcessing complete!\n")
	fmt.Printf("Successfully processed: %d files\n", processedCount)
	fmt.Printf("  Cropped: %d files\n", croppedCount)
	fmt.Printf("  Unchanged: %d files\n", unchangedCount)
	if errorCount > 0 {
		fmt.Printf("Errors encountered: %d files\n", errorCount)
	}
}
