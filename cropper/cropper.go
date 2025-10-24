package cropper

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
)

// CropImage crops an image from a specified corner by a given percentage
func CropImage(inputPath, outputPath, corner string, percent float64) error {
	// Open the input file
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer file.Close()

	// Decode the image
	img, err := jpeg.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Get image dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate crop dimensions
	cropWidth := int(float64(width) * (percent / 100.0))
	cropHeight := int(float64(height) * (percent / 100.0))

	// Calculate new dimensions after cropping
	newWidth := width - cropWidth
	newHeight := height - cropHeight

	// Determine the cropping rectangle based on the corner
	var cropRect image.Rectangle

	switch corner {
	case "tl": // Top-left: remove from top and left
		cropRect = image.Rect(cropWidth, cropHeight, width, height)
	case "tr": // Top-right: remove from top and right
		cropRect = image.Rect(0, cropHeight, newWidth, height)
	case "bl": // Bottom-left: remove from bottom and left
		cropRect = image.Rect(cropWidth, 0, width, newHeight)
	case "br": // Bottom-right: remove from bottom and right
		cropRect = image.Rect(0, 0, newWidth, newHeight)
	default:
		return fmt.Errorf("invalid corner: %s", corner)
	}

	// Create a new image with the cropped portion
	croppedImg := image.NewRGBA(image.Rect(0, 0, cropRect.Dx(), cropRect.Dy()))

	// Copy the cropped portion to the new image
	for y := cropRect.Min.Y; y < cropRect.Max.Y; y++ {
		for x := cropRect.Min.X; x < cropRect.Max.X; x++ {
			croppedImg.Set(x-cropRect.Min.X, y-cropRect.Min.Y, img.At(x, y))
		}
	}

	// Create the output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Encode and save the cropped image as JPEG
	options := &jpeg.Options{Quality: 95}
	if err := jpeg.Encode(outFile, croppedImg, options); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	return nil
}
