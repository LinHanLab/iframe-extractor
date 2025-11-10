# iframe-extractor

A lightweight Go tool that extracts iframe tags from web pages using headless Chrome. Perfect for analyzing embedded content, scraping iframe sources, or auditing third-party integrations.

## Features

- **Headless Browser**: Uses Chrome DevTools Protocol via [chromedp](https://github.com/chromedp/chromedp) to load pages fully, including dynamically rendered content
- **Concurrent Processing**: Process multiple URLs in parallel with configurable worker pools
- **Smart Waiting**: Random sleep intervals ensure dynamic content has time to load
- **Flexible Input**: Accept URLs via command line arguments or input file
- **Deduplication**: Automatically filters duplicate iframe sources
- **Resume Support**: Skip already processed URLs to avoid redundant work
- **Customizable Output**: Configure output directory and choose to save all iframes or just the first one

## Installation

### Prerequisites

- Go 1.25.0 or later
- Chrome/Chromium browser installed on your system

### Build from Source

```bash
git clone <repository-url>
cd iframe-extractor
go mod download
go build
```

## Usage

### Basic Usage

Extract iframes from a single URL:

```bash
./iframe-extractor https://example.com
```

Process multiple URLs:

```bash
./iframe-extractor https://example.com https://test.com https://site.org
```

### Using a URL File

Create a text file with URLs (one per line):

```text
# urls.txt
https://example.com
https://test.com

# Comments and blank lines are ignored
https://another-site.org
```

Process the file:

```bash
./iframe-extractor --urls-file urls.txt
```

### Configuration Options

| Flag | Default | Description |
|------|---------|-------------|
| `--urls-file` | - | Path to file containing URLs (one per line) |
| `--skip-already-exists-files` | `true` | Skip URLs whose output files already exist |
| `--max-workers` | `3` | Number of concurrent workers for parallel processing |
| `--timeout-seconds` | `70` | Browser timeout in seconds for loading each page |
| `--min-sleep-seconds` | `10` | Minimum wait time after page load for content to render |
| `--max-sleep-seconds` | `25` | Maximum wait time after page load for content to render |
| `--save-all-iframes` | `true` | Save all iframes found (false = save only first iframe) |
| `--output-dir` | `__output_iframes` | Directory where extracted iframe HTML files will be saved |

### Examples

Extract only the first iframe from each page:

```bash
./iframe-extractor --save-all-iframes=false https://example.com
```

Process with 10 concurrent workers and shorter timeouts:

```bash
./iframe-extractor --max-workers 10 --timeout-seconds 30 --urls-file urls.txt
```

Custom wait times for slow-loading dynamic content:

```bash
./iframe-extractor --min-sleep-seconds 15 --max-sleep-seconds 30 https://example.com
```

Save to a custom directory:

```bash
./iframe-extractor --output-dir my_iframes https://example.com
```

Reprocess all URLs (don't skip existing files):

```bash
./iframe-extractor --skip-already-exists-files=false --urls-file urls.txt
```

## How It Works

1. **Parse Input**: Reads URLs from command line arguments or input file
2. **Filter URLs**: Optionally skips URLs that have already been processed
3. **Concurrent Processing**: Spawns worker goroutines to process URLs in parallel
4. **Browser Automation**: For each URL:
   - Launches a headless Chrome instance
   - Navigates to the URL
   - Waits for the page body to be ready
   - Sleeps for a random duration (between min/max seconds) to allow dynamic content to load
   - Extracts iframe tags using regex pattern matching on the HTML
   - Filters out `about:blank` iframes and duplicates
5. **Save Results**: Creates HTML files containing the extracted iframe tags
6. **Report**: Displays progress and any errors encountered

## Output Format

Extracted iframes are saved as simple HTML files:

```html
<!DOCTYPE html>
<html>
<body>
<iframe src="https://example.com/embed/video123">
<iframe src="https://another-site.com/widget">
</body>
</html>
```

### Output Filenames

Filenames are derived from the URL's last path segment:
- `https://example.com/page/video` → `video.html`
- `https://example.com/page/` → `page.html`
- `https://example.com/` → `index.html`
- Query parameters are removed: `page?id=123` → `page.html`

## Dependencies

- [chromedp/chromedp](https://github.com/chromedp/chromedp) - Chrome DevTools Protocol driver for Go
