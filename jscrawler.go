package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/rix4uni/jscrawler/banner"
	flag "github.com/spf13/pflag"
	"golang.org/x/net/html"
)

type Config struct {
	timeout  int
	complete bool
	output   string
	verbose  bool
	silent   bool
	threads  int
	version  bool
}

func main() {
	config := Config{}

	flag.IntVar(&config.timeout, "timeout", 15, "Timeout (in seconds) for http client")
	flag.BoolVar(&config.complete, "complete", false, "Get Complete URL (default false)")
	flag.StringVarP(&config.output, "output", "o", "", "Output file to save results")
	flag.IntVarP(&config.threads, "threads", "t", 50, "Number of threads to use")
	flag.BoolVar(&config.silent, "silent", false, "Silent mode.")
	flag.BoolVar(&config.version, "version", false, "Print the version of the tool and exit.")
	flag.BoolVar(&config.verbose, "verbose", false, "Enable verbose output for debugging purposes.")

	// Parse the flags
	flag.Parse()

	if config.version {
		banner.PrintBanner()
		banner.PrintVersion()
		return
	}

	if !config.silent {
		banner.PrintBanner()
	}

	// Read URLs from stdin
	urls := readURLsFromStdin()

	if len(urls) == 0 {
		return
	}

	// Process URLs concurrently
	processURLsConcurrently(urls, config)
}

func readURLsFromStdin() []string {
	var urls []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url != "" {
			urls = append(urls, url)
		}
	}
	return urls
}

func processURLsConcurrently(urls []string, config Config) {
	// Create a semaphore to limit concurrent requests
	sem := make(chan struct{}, config.threads)
	var wg sync.WaitGroup

	// Create output file mutex if output is specified
	var outputMutex sync.Mutex
	var outputFile *os.File
	var err error

	if config.output != "" {
		outputFile, err = os.OpenFile(config.output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			if config.verbose {
				fmt.Fprintf(os.Stderr, "Error opening output file: %v\n", err)
			}
			return
		}
		defer outputFile.Close()
	}

	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			processURL(u, config, outputFile, &outputMutex)
		}(url)
	}

	wg.Wait()
}

func processURL(baseURL string, config Config, outputFile *os.File, outputMutex *sync.Mutex) {
	baseURL = strings.TrimSpace(baseURL)

	if config.verbose {
		fmt.Fprintf(os.Stderr, "Processing URL: %s\n", baseURL)
	}

	// Create HTTP client with timeout and TLS config
	client := &http.Client{
		Timeout: time.Duration(config.timeout) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Get(baseURL)
	if err != nil {
		if config.verbose {
			if os.IsTimeout(err) {
				fmt.Fprintf(os.Stderr, "Timeout occurred while fetching: %s\n", baseURL)
			} else if strings.Contains(err.Error(), "connection refused") {
				fmt.Fprintf(os.Stderr, "A connection error occurred. Please check your internet connection: %s\n", baseURL)
			} else {
				fmt.Fprintf(os.Stderr, "An error occurred: %v\n", err)
			}
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if config.verbose {
			fmt.Fprintf(os.Stderr, "HTTP Error %d: %s\n", resp.StatusCode, baseURL)
		}
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if config.verbose {
			fmt.Fprintf(os.Stderr, "Error reading response body: %v\n", err)
		}
		return
	}

	extractJSLinks(baseURL, string(body), config, outputFile, outputMutex)
}

func extractJSLinks(baseURL string, htmlContent string, config Config, outputFile *os.File, outputMutex *sync.Mutex) {
	jsLinks := make(map[string]bool)

	// Extract links using regex
	re := regexp.MustCompile(`['"]([^'"]*\.js[^'"]*)['"]`)
	matches := re.FindAllStringSubmatch(htmlContent, -1)
	for _, match := range matches {
		if len(match) > 1 {
			link := strings.Trim(match[1], `'"`)
			if !strings.Contains(link, ".json") {
				if config.complete {
					fullURL := resolveURL(baseURL, link)
					jsLinks[fullURL] = true
				} else {
					jsLinks[link] = true
				}
			}
		}
	}

	// Parse HTML and extract script src attributes
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err == nil {
		extractScriptTags(doc, baseURL, config.complete, jsLinks)
	}

	// Output the collected links
	for jsLink := range jsLinks {
		fmt.Println(jsLink)
		saveOutput(outputFile, jsLink, outputMutex)
	}
}

func extractScriptTags(n *html.Node, baseURL string, complete bool, jsLinks map[string]bool) {
	if n.Type == html.ElementNode && n.Data == "script" {
		var src string
		scriptType := ""

		for _, attr := range n.Attr {
			if attr.Key == "src" {
				src = attr.Val
			}
			if attr.Key == "type" {
				scriptType = strings.ToLower(attr.Val)
			}
		}

		if src != "" && (scriptType == "text/javascript" || scriptType == "") {
			if complete {
				fullURL := resolveURL(baseURL, src)
				jsLinks[fullURL] = true
			} else {
				jsLinks[src] = true
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractScriptTags(c, baseURL, complete, jsLinks)
	}
}

func resolveURL(baseURL, relativeURL string) string {
	// Handle absolute URLs
	if strings.HasPrefix(relativeURL, "http://") || strings.HasPrefix(relativeURL, "https://") {
		return relativeURL
	}

	// Handle protocol-relative URLs
	if strings.HasPrefix(relativeURL, "//") {
		if strings.HasPrefix(baseURL, "https://") {
			return "https:" + relativeURL
		}
		return "http:" + relativeURL
	}

	// Parse base URL
	base := strings.TrimSuffix(baseURL, "/")

	// Handle root-relative URLs
	if strings.HasPrefix(relativeURL, "/") {
		// Extract scheme and host from base URL
		schemeEnd := strings.Index(base, "://")
		if schemeEnd == -1 {
			return relativeURL
		}
		hostStart := schemeEnd + 3
		hostEnd := strings.Index(base[hostStart:], "/")
		if hostEnd == -1 {
			return base + relativeURL
		}
		return base[:hostStart+hostEnd] + relativeURL
	}

	// Handle relative URLs
	lastSlash := strings.LastIndex(base, "/")
	if lastSlash > strings.Index(base, "://")+2 {
		base = base[:lastSlash]
	}
	return base + "/" + relativeURL
}

func saveOutput(outputFile *os.File, url string, outputMutex *sync.Mutex) {
	if outputFile != nil {
		outputMutex.Lock()
		defer outputMutex.Unlock()
		outputFile.WriteString(url + "\n")
	}
}
