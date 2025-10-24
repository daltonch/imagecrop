# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`imagecrop` is a Go CLI tool that batch processes JPEG images by cropping them from a specified corner by a given percentage. It recursively processes all JPG/JPEG files in an input directory and saves the cropped versions to an output directory.

## Build and Run

Build the binary:
```bash
go mod tidy
go build -o imagecrop
```

Run the tool:
```bash
./imagecrop --input <input_dir> --corner <tl|tr|bl|br> --percent <0-100> [--output <output_dir>]
```

Example:
```bash
./imagecrop --input ./photos --corner tl --percent 10 --output ./cropped
```

## Architecture

The codebase follows a simple two-layer architecture:

1. **main.go**: CLI interface layer
   - Handles flag parsing and validation
   - Validates input/output directories
   - Walks the input directory tree using `filepath.WalkDir`
   - Filters for JPG/JPEG files
   - Reports processing summary

2. **cropper/cropper.go**: Image processing logic
   - `CropImage()` function handles the core cropping algorithm
   - Opens and decodes JPEG images using Go's `image/jpeg` package
   - Calculates crop dimensions based on percentage
   - Determines crop rectangle based on corner parameter (tl/tr/bl/br)
   - Manually copies pixels to create the cropped image
   - Encodes output as JPEG with 95% quality

The corner cropping logic works by removing pixels from two edges:
- `tl` (top-left): removes from top and left edges
- `tr` (top-right): removes from top and right edges
- `bl` (bottom-left): removes from bottom and left edges
- `br` (bottom-right): removes from bottom and right edges