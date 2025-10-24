# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`imagecrop` is a Go CLI tool that intelligently crops JPEG and PNG images based on brightness analysis. It detects non-uniform lighting (darker or brighter edges) and progressively crops edges to achieve uniform brightness. Images that are already uniformly lit are copied unchanged.

## Build and Run

Build the binary:
```bash
go mod tidy
go build -o imagecrop
```

Run the tool:
```bash
./imagecrop --input <input_dir> [--output <output_dir>] [--tolerance <0-100>] [--max-crop <0-100>]
```

Example:
```bash
./imagecrop --input ./photos --tolerance 15 --max-crop 30 --output ./cropped
```

## CLI Flags

- `--input` (required): Input directory containing image files (JPEG/JPG/PNG)
- `--output` (optional): Output directory, default: "cropped"
- `--tolerance` (optional): Brightness variation tolerance percentage (0-100), default: 15
- `--max-crop` (optional): Maximum crop percentage per dimension (0-100), default: 30
- `--threads` (optional): Number of concurrent processing threads, default: 4

## Architecture

The codebase follows a two-layer architecture:

### 1. main.go - CLI Interface Layer
- Parses and validates command-line flags
- Validates input/output directories
- Recursively walks the input directory using `filepath.WalkDir` to collect jobs
- Filters for image files (JPG/JPEG/PNG)
- **Multi-threaded Processing**:
  - Uses worker pool pattern with configurable number of threads
  - Job channel distributes work to concurrent workers
  - Each worker processes images with unique temp files
  - Thread-safe counters using `sync.Mutex`
  - Thread-safe console output using separate mutex
- Renames output files based on crop result:
  - Appends "_cropped" suffix if image was cropped
  - Uses original filename if unchanged
- Reports detailed summary (cropped count, unchanged count, errors)

### 2. cropper/cropper.go - Brightness Analysis and Cropping Logic

**Key Types:**
- `CropResult`: Contains `WasCropped` bool and `Message` string

**Main Function:**
- `CropImage(inputPath, outputPath, tolerance, maxCropPercent)`: Main entry point, returns `*CropResult`

**Algorithm Flow:**
1. Decode image (JPEG or PNG) using `image.Decode()`
2. Check if already uniform using `isUniform()`
3. If uniform: copy unchanged via `copyImage()`
4. If not uniform: call `findUniformCrop()` to iteratively crop edges
5. Save result in original format (JPEG at 95% quality or PNG)

**Brightness Analysis:**
- `calculateBrightness()`: Uses standard luminance formula Y = 0.299R + 0.587G + 0.114B
- `calculateRegionBrightness()`: Calculates average brightness for a rectangular region
- `isUniform()`: Samples 10% bands from each edge (top, bottom, left, right) and compares against **center region brightness** (inner 60% of image), not overall average. This prevents large dark/bright edge regions from skewing the reference.

**Progressive Cropping Algorithm (`findUniformCrop`):**
1. Calculate max pixels that can be cropped based on `maxCropPercent`
2. Start with full image bounds
3. Iterate up to `max(width, height)/2` times (e.g., 1920 iterations for 3840px wide images):
   - Check if current crop is uniform (within tolerance)
   - If uniform: return current crop rectangle
   - If max crop limit reached: return current crop
   - Calculate **center region brightness** (inner 60% of current crop)
   - Sample 5% bands from each edge
   - Calculate brightness deviation of each edge from center
   - Identify edge with maximum deviation
   - Crop approximately 1% of dimension (avg of width+height / 200) from that edge
   - Repeat
4. Return final crop rectangle

The algorithm progressively removes the "worst" edge (most deviation from center) in ~1% chunks until uniformity is achieved or limits are reached. The center-weighted approach and aggressive cropping make it effective for images with large non-uniform regions.

## Output Behavior

- Cropped images: `{original_name}_cropped.{ext}`
- Unchanged images: `{original_name}.{ext}` (no suffix)
- All images output to the specified output directory
