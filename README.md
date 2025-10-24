# imagecrop

A command-line tool for batch cropping JPEG images from a specified corner by a given percentage.

## Description

`imagecrop` recursively processes all JPEG files in a directory, cropping them from a specified corner (top-left, top-right, bottom-left, or bottom-right) by a given percentage. The cropped images are saved to an output directory while preserving the original filenames.

## Installation

Make sure you have Go installed (version 1.25.3 or later), then build the binary:

```bash
go mod tidy
go build -o imagecrop
```

## Usage

```bash
./imagecrop --input <input_dir> --corner <corner> --percent <percentage> [--output <output_dir>]
```

### Required Flags

- `--input`: Input directory containing JPG/JPEG files
- `--corner`: Corner to crop from. Valid values:
  - `tl` - Top-left (removes from top and left edges)
  - `tr` - Top-right (removes from top and right edges)
  - `bl` - Bottom-left (removes from bottom and left edges)
  - `br` - Bottom-right (removes from bottom and right edges)
- `--percent`: Percentage to crop (must be between 0 and 100)

### Optional Flags

- `--output`: Output directory for cropped images (default: `cropped`)

## Examples

Crop 10% from the top-left corner of all images in the `photos` directory:
```bash
./imagecrop --input ./photos --corner tl --percent 10
```

Crop 15% from the bottom-right corner and save to a custom output directory:
```bash
./imagecrop --input ./images --corner br --percent 15 --output ./processed
```

Crop 20% from the top-right corner:
```bash
./imagecrop --input ./vacation_pics --corner tr --percent 20 --output ./cropped_pics
```

## How It Works

The tool:
1. Recursively scans the input directory for all JPG/JPEG files
2. For each image, calculates the crop dimensions based on the percentage
3. Removes pixels from two edges based on the specified corner:
   - `tl`: Removes from top and left
   - `tr`: Removes from top and right
   - `bl`: Removes from bottom and left
   - `br`: Removes from bottom and right
4. Saves the cropped image to the output directory with the same filename
5. Reports a summary of successfully processed files and any errors

Cropped images are saved as JPEG files with 95% quality.

## Output

The tool provides real-time feedback as it processes files:
```
Processing: image1.jpg
  Saved to: cropped/image1.jpg
Processing: image2.jpg
  Saved to: cropped/image2.jpg

Processing complete!
Successfully processed: 2 files
```

If any errors occur, they will be reported but processing will continue with remaining files.