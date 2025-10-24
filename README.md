# imagecrop

An intelligent command-line tool for automatically cropping JPEG and PNG images based on brightness analysis to achieve uniform lighting.

## Description

`imagecrop` analyzes the brightness distribution of images and intelligently crops darker or brighter edges to produce uniformly lit results. The tool recursively processes all image files (JPEG/PNG) in a directory, automatically detecting which images need cropping and which are already uniform.

### Key Features

- **Intelligent brightness analysis**: Uses center-weighted luminance calculation to detect lighting variations
- **Progressive edge cropping**: Iteratively removes the worst edges until uniform lighting is achieved
- **Multi-threaded processing**: Concurrent image processing for faster batch operations
- **Configurable tolerance**: Control how strict the uniformity requirement is
- **Safe cropping limits**: Set maximum crop percentage to prevent over-cropping
- **Smart file naming**: Cropped images get "_cropped" suffix; unchanged images keep original names
- **Batch processing**: Automatically processes entire directories recursively

## Installation

Make sure you have Go installed (version 1.25.3 or later), then build the binary:

```bash
go mod tidy
go build -o imagecrop
```

## Usage

```bash
./imagecrop --input <input_dir> [options]
```

### Required Flags

- `--input`: Input directory containing image files (JPEG/JPG/PNG)

### Optional Flags

- `--output`: Output directory for processed images (default: `cropped`)
- `--tolerance`: Brightness variation tolerance percentage, 0-100 (default: `15`)
  - Lower values = stricter uniformity requirement = more aggressive cropping
  - Higher values = more lenient = less cropping
- `--max-crop`: Maximum percentage to crop from any dimension, 0-100 (default: `30`)
  - Prevents over-cropping that would make images too small
  - Applied per dimension (width and height independently)
- `--threads`: Number of concurrent processing threads (default: `4`)
  - Higher values = faster processing for large batches
  - Recommended: set to number of CPU cores for best performance

## Examples

Process images with default settings (15% tolerance, 30% max crop, 4 threads):
```bash
./imagecrop --input ./photos
```

Use strict uniformity requirement with conservative cropping:
```bash
./imagecrop --input ./images --tolerance 10 --max-crop 20 --output ./processed
```

More lenient tolerance for naturally varied lighting:
```bash
./imagecrop --input ./vacation_pics --tolerance 25 --max-crop 40
```

Process large batch with 8 threads for maximum performance:
```bash
./imagecrop --input ./raw_photos --output ./corrected_photos --threads 8
```

Single-threaded processing for debugging:
```bash
./imagecrop --input ./photos --threads 1
```

## How It Works

The tool uses an intelligent brightness analysis algorithm with center-weighted reference:

1. **Image Analysis**: Each image is scanned to calculate brightness distribution using the standard luminance formula: `Y = 0.299*R + 0.587*G + 0.114*B`

2. **Center-Weighted Uniformity Check**:
   - Calculates the brightness of the center 60% of the image as the reference
   - Samples edge regions (10% bands from top, bottom, left, right)
   - Compares edge brightness to center brightness (not overall average)
   - This prevents large dark/bright edge regions from skewing the reference

3. **Progressive Cropping**: If non-uniform:
   - Identifies which edge has the greatest brightness deviation from center
   - Removes approximately 1% of the dimension from that edge
   - Recalculates center brightness with the new crop
   - Repeats until uniform or max-crop limit reached

4. **Smart Output**:
   - Already uniform images → copied unchanged with original filename
   - Cropped images → saved with "_cropped" appended to filename
   - Example: `photo.jpg` becomes `photo_cropped.jpg`

5. **Multi-Threaded Batch Processing**:
   - Processes multiple images concurrently using worker threads
   - All image files (JPEG/PNG) in the input directory and subdirectories
   - Thread-safe output and statistics

## Output

The tool provides real-time feedback as it processes files:

```
Found 92 images to process using 4 threads...

Processing: sunset.jpg
Processing: portrait.jpg
Processing: landscape.jpg
Processing: mountain.jpg
  cropped 12.3% of image area -> sunset_cropped.jpg
  already uniform, copied unchanged -> portrait.jpg
  cropped 8.5% of image area -> landscape_cropped.jpg
  cropped 45.2% of image area -> mountain_cropped.jpg

Processing complete!
Successfully processed: 4 files
  Cropped: 3 files
  Unchanged: 1 files
```

Note: With multi-threading, processing and completion messages may appear interleaved as multiple images are processed concurrently.

## Understanding the Algorithm

**Center-Weighted Reference**: The algorithm compares edge brightness to the center region (inner 60%)
- This approach works well with images that have large dark/bright edge regions
- The center region serves as the "target" brightness for the entire image
- Prevents edge regions from skewing the reference brightness

**Tolerance Parameter**: Controls how much brightness variation is acceptable
- At 15% tolerance: edges can be up to 15% brighter/darker than center before cropping
- Lower tolerance (e.g., 10%) = more uniform results but crops more aggressively
- Higher tolerance (e.g., 20%) = preserves more of the image but allows more variation

**Max-Crop Parameter**: Safety limit to prevent excessive cropping
- At 30% max-crop: tool won't remove more than 30% of width or height
- Ensures you don't end up with tiny images from aggressive cropping
- If uniformity can't be achieved within the limit, it stops and saves the best attempt

**Performance**: Multi-threading provides significant speedup
- 4 threads (default): Good for most systems, balanced performance
- 8+ threads: Recommended for large batches on high-core systems
- Scaling depends on CPU cores and disk I/O speed

## Use Cases

- **Product photography**: Remove dark corners or edges from product shots
- **Document scanning**: Clean up unevenly lit scanned documents
- **Vignette removal**: Automatically crop darkened edges from photos
- **Batch correction**: Process hundreds of images with uneven lighting automatically

## Limitations

- Only processes JPEG/JPG and PNG files (not TIFF, GIF, etc.)
- Cropping is destructive - always keep original files
- Very complex lighting scenarios may not achieve perfect uniformity
- Processing speed depends on image size and aggressiveness of cropping needed
- PNG files preserve transparency but may result in larger file sizes than JPEG

# Binaries
- Download them from Actions