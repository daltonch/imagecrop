package cropper

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
)

// CropResult contains information about the cropping operation
type CropResult struct {
	WasCropped bool
	Message    string
}

// CropImage analyzes an image's brightness and crops edges that are significantly
// darker or brighter than the rest of the image to achieve uniform lighting
func CropImage(inputPath, outputPath string, tolerance, maxCropPercent float64) (*CropResult, error) {
	// Open the input file
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open input file: %w", err)
	}
	defer file.Close()

	// Decode the image (supports JPEG and PNG)
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Check if image is already uniform
	if isUniform(img, bounds, tolerance) {
		// Copy unchanged
		return copyImage(inputPath, outputPath)
	}

	// Perform iterative cropping to achieve uniform brightness
	cropRect, err := findUniformCrop(img, bounds, tolerance, maxCropPercent)
	if err != nil {
		return nil, err
	}

	// Check if we ended up cropping anything
	if cropRect.Dx() == width && cropRect.Dy() == height {
		// No crop was possible while staying within limits
		return copyImage(inputPath, outputPath)
	}

	// Create and save the cropped image
	croppedImg := image.NewRGBA(image.Rect(0, 0, cropRect.Dx(), cropRect.Dy()))
	for y := cropRect.Min.Y; y < cropRect.Max.Y; y++ {
		for x := cropRect.Min.X; x < cropRect.Max.X; x++ {
			croppedImg.Set(x-cropRect.Min.X, y-cropRect.Min.Y, img.At(x, y))
		}
	}

	// Save the cropped image
	outFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Encode based on detected format or output file extension
	outputExt := strings.ToLower(filepath.Ext(outputPath))
	if outputExt == ".png" || format == "png" {
		if err := png.Encode(outFile, croppedImg); err != nil {
			return nil, fmt.Errorf("failed to encode PNG image: %w", err)
		}
	} else {
		// Default to JPEG
		options := &jpeg.Options{Quality: 95}
		if err := jpeg.Encode(outFile, croppedImg, options); err != nil {
			return nil, fmt.Errorf("failed to encode JPEG image: %w", err)
		}
	}

	cropPercent := (1.0 - float64(cropRect.Dx()*cropRect.Dy())/float64(width*height)) * 100
	return &CropResult{
		WasCropped: true,
		Message:    fmt.Sprintf("cropped %.1f%% of image area", cropPercent),
	}, nil
}

// copyImage copies an image file unchanged
func copyImage(inputPath, outputPath string) (*CropResult, error) {
	input, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file: %w", err)
	}

	if err := os.WriteFile(outputPath, input, 0644); err != nil {
		return nil, fmt.Errorf("failed to write output file: %w", err)
	}

	return &CropResult{
		WasCropped: false,
		Message:    "already uniform, copied unchanged",
	}, nil
}

// calculateBrightness calculates the perceived brightness of a color using luminance formula
func calculateBrightness(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	// Convert from 16-bit to 8-bit and apply standard luminance formula
	// Y = 0.299*R + 0.587*G + 0.114*B
	return 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
}

// calculateRegionBrightness calculates average brightness for a region
func calculateRegionBrightness(img image.Image, rect image.Rectangle) float64 {
	var sum float64
	count := 0

	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			sum += calculateBrightness(img.At(x, y))
			count++
		}
	}

	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

// isUniform checks if the image has uniform brightness within tolerance
func isUniform(img image.Image, bounds image.Rectangle, tolerance float64) bool {
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate center region brightness (inner 60% of image)
	// This prevents large dark edge regions from skewing the reference brightness
	centerMarginX := width / 5 // 20% margin on each side = 60% center
	centerMarginY := height / 5
	if centerMarginX < 1 {
		centerMarginX = 1
	}
	if centerMarginY < 1 {
		centerMarginY = 1
	}

	centerRect := image.Rect(
		bounds.Min.X+centerMarginX,
		bounds.Min.Y+centerMarginY,
		bounds.Max.X-centerMarginX,
		bounds.Max.Y-centerMarginY,
	)

	// Ensure center rect is valid
	if centerRect.Dx() <= 0 || centerRect.Dy() <= 0 {
		// Image too small, fall back to overall average
		centerRect = bounds
	}

	centerBrightness := calculateRegionBrightness(img, centerRect)

	// Sample size for edge analysis (10% of dimension)
	sampleWidth := width / 10
	if sampleWidth < 1 {
		sampleWidth = 1
	}
	sampleHeight := height / 10
	if sampleHeight < 1 {
		sampleHeight = 1
	}

	// Check top edge
	topRect := image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Min.Y+sampleHeight)
	topBrightness := calculateRegionBrightness(img, topRect)
	if math.Abs(topBrightness-centerBrightness)/centerBrightness*100 > tolerance {
		return false
	}

	// Check bottom edge
	bottomRect := image.Rect(bounds.Min.X, bounds.Max.Y-sampleHeight, bounds.Max.X, bounds.Max.Y)
	bottomBrightness := calculateRegionBrightness(img, bottomRect)
	if math.Abs(bottomBrightness-centerBrightness)/centerBrightness*100 > tolerance {
		return false
	}

	// Check left edge
	leftRect := image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Min.X+sampleWidth, bounds.Max.Y)
	leftBrightness := calculateRegionBrightness(img, leftRect)
	if math.Abs(leftBrightness-centerBrightness)/centerBrightness*100 > tolerance {
		return false
	}

	// Check right edge
	rightRect := image.Rect(bounds.Max.X-sampleWidth, bounds.Min.Y, bounds.Max.X, bounds.Max.Y)
	rightBrightness := calculateRegionBrightness(img, rightRect)
	if math.Abs(rightBrightness-centerBrightness)/centerBrightness*100 > tolerance {
		return false
	}

	return true
}

// findUniformCrop progressively crops edges to achieve uniform brightness
func findUniformCrop(img image.Image, bounds image.Rectangle, tolerance, maxCropPercent float64) (image.Rectangle, error) {
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate maximum pixels we can crop from each dimension
	maxCropWidth := int(float64(width) * maxCropPercent / 100.0)
	maxCropHeight := int(float64(height) * maxCropPercent / 100.0)

	// Start with full image
	cropRect := bounds

	// Iteratively crop edges that are non-uniform
	// Allow enough iterations for large images (e.g., 4K images may need 2000+ iterations)
	maxIterations := int(math.Max(float64(width), float64(height))) / 2
	if maxIterations < 100 {
		maxIterations = 100
	}

	for i := 0; i < maxIterations; i++ {
		// Check if current crop is uniform
		if isUniform(img, cropRect, tolerance) {
			return cropRect, nil
		}

		// Calculate current crop dimensions
		currentWidth := cropRect.Dx()
		currentHeight := cropRect.Dy()

		// Check if we've hit the crop limit
		croppedWidth := width - currentWidth
		croppedHeight := height - currentHeight

		if croppedWidth >= maxCropWidth && croppedHeight >= maxCropHeight {
			// Can't crop anymore
			return cropRect, nil
		}

		// Calculate center region brightness (inner 60% of current crop)
		// This prevents large dark edge regions from skewing the reference brightness
		centerMarginX := currentWidth / 5 // 20% margin on each side = 60% center
		centerMarginY := currentHeight / 5
		if centerMarginX < 1 {
			centerMarginX = 1
		}
		if centerMarginY < 1 {
			centerMarginY = 1
		}

		centerCropRect := image.Rect(
			cropRect.Min.X+centerMarginX,
			cropRect.Min.Y+centerMarginY,
			cropRect.Max.X-centerMarginX,
			cropRect.Max.Y-centerMarginY,
		)

		// Ensure center rect is valid
		var centerBrightness float64
		if centerCropRect.Dx() <= 0 || centerCropRect.Dy() <= 0 {
			// Image too small, fall back to current crop area
			centerBrightness = calculateRegionBrightness(img, cropRect)
		} else {
			centerBrightness = calculateRegionBrightness(img, centerCropRect)
		}

		// Sample size for edge detection (5% of current dimension)
		sampleWidth := currentWidth / 20
		if sampleWidth < 1 {
			sampleWidth = 1
		}
		sampleHeight := currentHeight / 20
		if sampleHeight < 1 {
			sampleHeight = 1
		}

		// Check each edge and find the one that deviates most
		edges := make(map[string]float64)

		// Top edge
		if croppedHeight < maxCropHeight {
			topRect := image.Rect(cropRect.Min.X, cropRect.Min.Y, cropRect.Max.X, cropRect.Min.Y+sampleHeight)
			topBrightness := calculateRegionBrightness(img, topRect)
			edges["top"] = math.Abs(topBrightness - centerBrightness)
		}

		// Bottom edge
		if croppedHeight < maxCropHeight {
			bottomRect := image.Rect(cropRect.Min.X, cropRect.Max.Y-sampleHeight, cropRect.Max.X, cropRect.Max.Y)
			bottomBrightness := calculateRegionBrightness(img, bottomRect)
			edges["bottom"] = math.Abs(bottomBrightness - centerBrightness)
		}

		// Left edge
		if croppedWidth < maxCropWidth {
			leftRect := image.Rect(cropRect.Min.X, cropRect.Min.Y, cropRect.Min.X+sampleWidth, cropRect.Max.Y)
			leftBrightness := calculateRegionBrightness(img, leftRect)
			edges["left"] = math.Abs(leftBrightness - centerBrightness)
		}

		// Right edge
		if croppedWidth < maxCropWidth {
			rightRect := image.Rect(cropRect.Max.X-sampleWidth, cropRect.Min.Y, cropRect.Max.X, cropRect.Max.Y)
			rightBrightness := calculateRegionBrightness(img, rightRect)
			edges["right"] = math.Abs(rightBrightness - centerBrightness)
		}

		// If no edges can be cropped, we're done
		if len(edges) == 0 {
			return cropRect, nil
		}

		// Find edge with maximum deviation
		var maxEdge string
		var maxDeviation float64
		for edge, deviation := range edges {
			if deviation > maxDeviation {
				maxDeviation = deviation
				maxEdge = edge
			}
		}

		// If max deviation is within tolerance, we're done
		if maxDeviation/centerBrightness*100 <= tolerance {
			return cropRect, nil
		}

		// Crop the edge with maximum deviation
		// Crop more aggressively (1% of dimension or at least 1 pixel) to speed up processing
		cropAmount := int(math.Max(1, float64(currentWidth+currentHeight)/200))

		switch maxEdge {
		case "top":
			cropRect.Min.Y += cropAmount
		case "bottom":
			cropRect.Max.Y -= cropAmount
		case "left":
			cropRect.Min.X += cropAmount
		case "right":
			cropRect.Max.X -= cropAmount
		}

		// Sanity check
		if cropRect.Dx() <= 0 || cropRect.Dy() <= 0 {
			return bounds, fmt.Errorf("crop would result in empty image")
		}
	}

	return cropRect, nil
}
