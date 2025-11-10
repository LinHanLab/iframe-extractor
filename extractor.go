package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

type Extractor struct {
	config Config
}

func NewExtractor(config Config) *Extractor {
	return &Extractor{
		config: config,
	}
}

type job struct {
	index int
	url   string
}

type result struct {
	url   string
	count int
	err   error
}

func (e *Extractor) Process(ctx context.Context) error {
	urls := e.getTargetURLs()
	if len(urls) == 0 {
		fmt.Println("All URLs skipped - no new processing needed")
		return nil
	}
	fmt.Printf("Processing %d URLs for iframe extraction with %d workers...\n", len(urls), e.config.MaxWorkers)

	jobs := make(chan job, len(urls))
	results := make(chan result, len(urls))

	var wg sync.WaitGroup
	for w := 0; w < e.config.MaxWorkers; w++ {
		wg.Add(1)
		go e.worker(ctx, jobs, results, &wg, len(urls))
	}

	for i, url := range urls {
		jobs <- job{index: i, url: url}
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var errorList []error
	for res := range results {
		if res.err != nil {
			errorList = append(errorList, res.err)
		}
	}

	if len(errorList) > 0 {
		fmt.Println("\nErrors encountered:")
		for _, err := range errorList {
			fmt.Printf("- %v\n", err)
		}
	}

	return errors.Join(errorList...)
}

func (e *Extractor) worker(ctx context.Context, jobs <-chan job, results chan<- result, wg *sync.WaitGroup, totalURLs int) {
	defer wg.Done()

	for j := range jobs {
		fmt.Printf("[%d/%d] Processing: %s\n", j.index+1, totalURLs, j.url)

		iframeData, err := e.extractIframes(ctx, j.url)
		if err != nil {
			results <- result{url: j.url, err: fmt.Errorf("%s: %v", j.url, err)}
			continue
		}

		if len(iframeData) == 0 {
			results <- result{url: j.url, err: fmt.Errorf("%s: no iframe tags found", j.url)}
			continue
		}

		filename := ConvertURLToFilename(j.url)
		if err := e.saveIframeFile(filename, iframeData); err != nil {
			results <- result{url: j.url, err: fmt.Errorf("%s: %v", j.url, err)}
		} else {
			outputPath := e.outputPath(j.url)
			fmt.Printf("  ✓ Saved %d iframes to %s\n", len(iframeData), outputPath)
			results <- result{url: j.url, count: len(iframeData)}
		}
	}
}

func (e *Extractor) getTargetURLs() []string {
	urls := make([]string, 0, len(e.config.URLs))
	for _, url := range e.config.URLs {
		outputPath := e.outputPath(url)
		_, err := os.Stat(outputPath)
		if err == nil && e.config.SkipAlreadyExistsFiles {
			fmt.Printf("[-] Skipping %s (already exists: %s)\n", url, outputPath)
			continue
		}
		urls = append(urls, url)
	}
	return urls
}

func (e *Extractor) outputPath(url string) string {
	filename := ConvertURLToFilename(url)
	return fmt.Sprintf("%s/%s.html", e.config.OutputDir, filename)
}

func (e *Extractor) extractIframes(ctx context.Context, url string) (iframeHTML []string, err error) {
	chromectx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	chromectx, cancel = context.WithTimeout(chromectx, e.config.Timeout)
	defer cancel()

	sleepRange := e.config.MaxSleepSeconds - e.config.MinSleepSeconds + 1
	randomSleep := time.Duration(e.config.MinSleepSeconds+rand.Intn(sleepRange)) * time.Second

	err = chromedp.Run(chromectx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(randomSleep),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Try direct HTML parsing first - this works better for dynamically loaded content
			var htmlContent string
			if err := chromedp.OuterHTML("html", &htmlContent).Do(ctx); err == nil {
				re := regexp.MustCompile(`<iframe[^>]*src=["']([^"']*)["'][^>]*>`)
				matches := re.FindAllStringSubmatch(htmlContent, -1)

				seenSources := make(map[string]bool)
				for _, match := range matches {
					if len(match) > 1 && match[1] != "" {
						src := match[1]
						if src != "about:blank" && !seenSources[src] {
							seenSources[src] = true
							iframeHTML = append(iframeHTML, fmt.Sprintf(`<iframe src="%s">`, src))

							if !e.config.SaveAllIframes && len(iframeHTML) == 1 {
								break
							}
						}
					}
				}
			}
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("headless browser error: %v", err)
	}

	return iframeHTML, nil
}

func (e *Extractor) saveIframeFile(filename string, iframeData []string) error {
	if err := os.MkdirAll(e.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	outputPath := fmt.Sprintf("%s/%s.html", e.config.OutputDir, filename)
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	template := `<!DOCTYPE html>
<html>
<body>
`
	for _, iframe := range iframeData {
		template += iframe + "\n"
	}
	template += `</body>
</html>`

	_, err = file.WriteString(template)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}
