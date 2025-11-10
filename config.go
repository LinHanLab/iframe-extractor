package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type Config struct {
	URLs []string

	SkipAlreadyExistsFiles bool
	MaxWorkers             int
	Timeout                time.Duration
	MinSleepSeconds        int
	MaxSleepSeconds        int

	SaveAllIframes bool
	OutputDir      string
}

const (
	defaultSkipAlreadyExistsFiles = true
	defaultMaxWorkers             = 3
	defaultTimeout                = 70
	defaultMinSleep               = 10
	defaultMaxSleep               = 25

	defaultSaveAllIframes = true
	defaultOutputDir      = "__output_iframes"
)

var (
	urlsFile = flag.String("urls-file", "", "Path to file containing URLs (one per line; empty lines and # comments are ignored)")

	skipAlreadyExistsFiles = flag.Bool("skip-already-exists-files", defaultSkipAlreadyExistsFiles, "Skip URLs whose output HTML files already exist in output directory")
	maxWorkers             = flag.Int("max-workers", defaultMaxWorkers, "Number of concurrent workers for parallel URL processing")
	timeoutSeconds         = flag.Int("timeout-seconds", defaultTimeout, "Browser timeout in seconds for loading each page")
	minSleepSeconds        = flag.Int("min-sleep-seconds", defaultMinSleep, "Minimum wait time (seconds) after page loads for content to fully render")
	maxSleepSeconds        = flag.Int("max-sleep-seconds", defaultMaxSleep, "Maximum wait time (seconds) after page loads for content to fully render")
	saveAllIframes         = flag.Bool("save-all-iframes", defaultSaveAllIframes, "Save all iframes found on page (false = save only first iframe)")
	outputDir              = flag.String("output-dir", defaultOutputDir, "Directory where extracted iframe HTML files will be saved")
)

func init() {
	flag.Usage = customUsage
}

const usageHeader = `iframe-extractor - Extract iframe tags from web pages using headless Chrome

USAGE:
  iframe-extractor [OPTIONS] URL [URL...]
  iframe-extractor [OPTIONS] --urls-file FILE

DESCRIPTION:
  This tool visits web pages using a headless Chrome browser, waits for content
  to load, and extracts all iframe tags. The extracted iframes are saved to HTML
  files in the output directory. Supports concurrent processing of multiple URLs.

INPUT MODES:
  1. Direct URLs: Pass one or more URLs as command-line arguments
  2. File input:  Use --urls-file to read URLs from a file (one per line)
                  Empty lines and lines starting with # are ignored

OPTIONS:
`

const usageExamples = `
EXAMPLES:
  # Extract iframes from a single URL
  iframe-extractor https://example.com

  # Process multiple URLs with custom workers and timeout
  iframe-extractor --max-workers 5 --timeout-seconds 30 https://example.com https://test.com

  # Read URLs from a file and save to custom directory
  iframe-extractor --urls-file urls.txt --output-dir my_iframes

  # Extract only the first iframe from each page
  iframe-extractor --save-all-iframes=false --urls-file urls.txt

  # Process with custom wait times for dynamic content
  iframe-extractor --min-sleep-seconds 5 --max-sleep-seconds 15 https://example.com

OUTPUT:
  HTML files are saved to the output directory with filenames derived from the
  URL path. Each file contains a simple HTML document with the extracted iframe
  tags. Existing files are skipped by default (use --skip-already-exists-files=false
  to override).

`

func customUsage() {
	fmt.Fprint(os.Stderr, usageHeader)
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, usageExamples)
}

func ProcessInput() Config {
	flag.Parse()

	args := flag.Args()
	urlsFilePath := *urlsFile

	if len(args) > 0 && urlsFilePath != "" {
		fmt.Println("cannot use both direct URLs and --input file simultaneously")
		flag.Usage()
		os.Exit(1)
	}

	var urls []string
	if urlsFilePath != "" {
		urls = readURLsFromFile(urlsFilePath)
	} else {
		if len(args) == 0 {
			fmt.Println("No URLs provided")
			flag.Usage()
			os.Exit(1)
		}
		urls = args
	}

	config := Config{
		URLs:                   urls,
		SkipAlreadyExistsFiles: *skipAlreadyExistsFiles,
		MaxWorkers:             *maxWorkers,
		Timeout:                time.Duration(*timeoutSeconds) * time.Second,
		MinSleepSeconds:        *minSleepSeconds,
		MaxSleepSeconds:        *maxSleepSeconds,
		SaveAllIframes:         *saveAllIframes,
		OutputDir:              *outputDir,
	}

	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal config: %v", err)
	}

	fmt.Println("=== Configuration ===")
	fmt.Println(string(configJSON))
	fmt.Println("=====================\n")

	return config
}

func readURLsFromFile(filename string) (urlList []string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("failed to open input file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		urlList = append(urlList, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("error reading file at line %d: %v", lineNum, err)
	}
	return urlList
}
