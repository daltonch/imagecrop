package main

import (
	"flag"
	"fmt"
	"imagecrop/cropper"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Define CLI flags
	inputDir := flag.String("input", "", "Input directory containing JPG files (required)")
	corner := flag.String("corner", "", "Corner to crop from: tl (top-left), tr (top-right), bl (bottom-left), br (bottom-right) (required)")
	percent := flag.Float64("percent", 0, "Percentage to crop (0-100) (required)")
	outputDir := flag.String("output", "cropped", "Output directory (default: cropped)")

	flag.Parse()

	// Validate required flags
	if *inputDir == "" {
		fmt.Println("Error: --input flag is required")
		flag.Usage()
		os.Exit(1)
	}

	if *corner == "" {
		fmt.Println("Error: --corner flag is required")
		flag.Usage()
		os.Exit(1)
	}

	if *percent <= 0 || *percent >= 100 {
		fmt.Println("Error: --percent must be between 0 and 100")
		flag.Usage()
		os.Exit(1)
	}

	// Validate corner value
	validCorners := map[string]bool{"tl": true, "tr": true, "bl": true, "br": true}
	if !validCorners[*corner] {
		fmt.Println("Error: --corner must be one of: tl, tr, bl, br")
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

	// Process all JPG files in the input directory
	processedCount := 0
	errorCount := 0

	err := filepath.WalkDir(*inputDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-JPG files
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".jpg" && ext != ".jpeg" {
			return nil
		}

		// Get the filename for the output
		filename := filepath.Base(path)
		outputPath := filepath.Join(*outputDir, filename)

		// Crop the image
		fmt.Printf("Processing: %s\n", filename)
		if err := cropper.CropImage(path, outputPath, *corner, *percent); err != nil {
			fmt.Printf("  Error: %v\n", err)
			errorCount++
			return nil // Continue processing other files
		}

		fmt.Printf("  Saved to: %s\n", outputPath)
		processedCount++
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	// Print summary
	fmt.Printf("\nProcessing complete!\n")
	fmt.Printf("Successfully processed: %d files\n", processedCount)
	if errorCount > 0 {
		fmt.Printf("Errors encountered: %d files\n", errorCount)
	}
}
