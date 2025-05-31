package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

// DataRecord represents a single record of scraped data
type DataRecord map[string]string

// ScraperConfig holds the configuration for the scraper
type ScraperConfig struct {
	URL           string
	OutputFormat  string
	OutputFile    string
	MaxRecords    int
	WaitTime      int
	Timeout       int // New field for timeout in seconds
	PaginationCSS string
	RecordCSS     string
	Fields        map[string]string
}

func main() {
	// Parse command line flags
	url := flag.String("url", "https://www.coupang.com/np/categories/195756", "URL to scrape")
	fieldsMapping := flag.String("mapping", "input/coupang_fields.json", "Mapping of the fields to be extracted")
	outputFormat := flag.String("format", "json", "Output format (json or csv)")
	outputFile := flag.String("output", "coupang", "Output file name (without extension)")
	maxRecords := flag.Int("max", 100, "Maximum number of records to scrape")
	waitTime := flag.Int("wait", 2, "Timeout configuration")
	paginationCSS := flag.String("pagination", "a.next-page", "CSS selector for pagination element")
	recordCSS := flag.String("record", "li.baby-product", "CSS selector for record elements")
	scrapeTimeout := flag.Int("timeout", 45, "Timeout in seconds for chromedp operations") // New flag
	flag.Parse()

	mappings, err := os.ReadFile(*fieldsMapping)
	if err != nil {
		log.Fatalf("Error reading fields mapping file: %v", err)
		return
	}

	fields := make(map[string]string)
	err = json.Unmarshal(mappings, &fields)
	if err != nil {
		log.Fatalf("Error parsing fields mapping: %v", err)
		return
	}

	// Create scraper config
	config := ScraperConfig{
		URL:           *url,
		OutputFormat:  *outputFormat,
		OutputFile:    *outputFile,
		MaxRecords:    *maxRecords,
		WaitTime:      *waitTime,
		Timeout:       *scrapeTimeout, // Initialize new field
		PaginationCSS: *paginationCSS,
		RecordCSS:     *recordCSS,
		Fields:        fields,
	}

	var records []DataRecord
	// Run the scraper
	for i := 0; i < 3; i++ {
		var err error
		records, err = scrape(config)
		if err != nil || len(records) == 0 {
			time.Sleep(2 * time.Second)
			log.Printf("======Error scraping data: %v. Retrying...======", err)
		} else if len(records) > 0 {
			break
		}
	}

	// Save the data
	if err := saveData(records, config); err != nil {
		log.Fatalf("Error saving data: %v", err)
	}

	// Print summary
	printSummary(records)
}

// scrape uses chromedp for JavaScript rendering
func scrape(config ScraperConfig) ([]DataRecord, error) {
	var records []DataRecord

	// Create a new context with options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36"),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Create a new browser context
	// parentCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithDebugf(log.Printf))
	parentCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	baseConfigURL := config.URL
	currentURL := config.URL
	hasNextPage := true
	pageNum := 0
	if err := chromedp.Run(parentCtx); err != nil {
		log.Fatal("context timeout reached, attempting to perform actions", err)
	}

	for hasNextPage && len(records) < config.MaxRecords {
		fmt.Printf("Scraping page %d with headless browser: %s\n", pageNum, currentURL)

		// Navigate to the page and wait for content to load
		var outputHtml string

		ctx1, cancel2 := context.WithTimeout(parentCtx, time.Duration(config.Timeout)*time.Second)
		defer cancel2() // Ensure the context is canceled to release resources

		if err1 := chromedp.Run(ctx1,
			chromedp.Navigate(currentURL),
			chromedp.Sleep(2*time.Second),
			// Wait for product elements to be visible
			chromedp.WaitVisible(config.RecordCSS, chromedp.ByQuery),
			// Scroll down to load lazy-loaded content
			chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight)`, nil),
			chromedp.Sleep(8*time.Second),
			chromedp.OuterHTML("html", &outputHtml),
		); err1 != nil {
			log.Println("context timeout reached, attempting to perform actions", err1)

			ctx2, cancel := context.WithTimeout(parentCtx, time.Duration(15)*time.Second)
			defer cancel() // Ensure the context is canceled to release resources

			_ = chromedp.Run(ctx2,
				chromedp.OuterHTML("html", &outputHtml),
			)
		}
		if outputHtml == "" {
			return records, nil
		}

		// Parse the HTML
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(outputHtml))
		if err != nil {
			return nil, fmt.Errorf("error parsing HTML: %v", err)
		}

		// Extract records from the current page
		productCount := 0
		doc.Find(config.RecordCSS).Each(func(i int, s *goquery.Selection) {
			if len(records) >= config.MaxRecords {
				return
			}

			record := make(DataRecord)
			for field, selector := range config.Fields {
				var value string
				if strings.Contains(selector, "@attr:") {
					// Extract attribute
					parts := strings.Split(selector, "@attr:")
					baseSelector := parts[0]
					attrName := parts[1]
					value, _ = s.Find(baseSelector).Attr(attrName)

					// Handle relative URLs for images and links
					if (field == "image_url" || field == "product_url") && value != "" && !strings.HasPrefix(value, "http") {
						baseURL := getBaseURL(currentURL)
						value = baseURL + value
					}
				} else {
					// Extract text
					value = strings.TrimSpace(s.Find(selector).Text())
				}
				record[field] = value
			}

			// Only add record if it has a product name
			if record["product_name"] != "" {
				records = append(records, record)
				productCount++
			}
		})

		fmt.Printf("Found %d products on page %d\n", productCount, pageNum)

		// Check for pagination
		hasNextPage = false
		pagination := doc.Find(config.PaginationCSS)

		nextURL, exists := pagination.Attr("href")
		if exists && productCount > 0 && strings.Contains(nextURL, "page=") {
			// Handle relative URLs
			if !strings.HasPrefix(nextURL, "http") {
				baseURL := getBaseURL(currentURL)
				nextURL = baseURL + nextURL
			}
			currentURL = nextURL
			hasNextPage = true
			pageNum++
		} else if productCount > 0 {
			pageNum++
			if strings.Contains(baseConfigURL, "?") {
				currentURL = baseConfigURL + "&page="
			} else {
				currentURL = baseConfigURL + "?page="
			}
			currentURL += fmt.Sprint(pageNum)
			hasNextPage = true
		}

		// Rate limiting
		time.Sleep(time.Duration(config.WaitTime) * time.Second)
	}

	return records, nil
}

// getBaseURL extracts the base URL from a full URL
func getBaseURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) < 3 {
		return url
	}

	// Handle URLs with http/https
	if parts[0] == "http:" || parts[0] == "https:" {
		baseURL := parts[0] + "//" + parts[2] + "/"
		return baseURL
	}

	return url
}

// saveData saves the scraped data to a file in the specified format
func saveData(records []DataRecord, config ScraperConfig) error {
	fileName := config.OutputFile

	switch strings.ToLower(config.OutputFormat) {
	case "json":
		fileName += ".json"
		return saveAsJSON(records, fileName)
	case "csv":
		fileName += ".csv"
		return saveAsCSV(records, fileName)
	default:
		return fmt.Errorf("unsupported output format: %s", config.OutputFormat)
	}
}

// saveAsJSON saves the data as JSON
func saveAsJSON(records []DataRecord, fileName string) error {
	// Fix URLs before saving
	for i := range records {
		// Fix product_url
		if url, ok := records[i]["product_url"]; ok && url != "" {
			// Fix multiple slashes
			url = strings.Replace(url, "//vp", "/vp", -1)
			// Fix escaped characters - using a more direct approach
			url = strings.ReplaceAll(url, "\\u0026", "&")
			url = strings.ReplaceAll(url, "\u0026", "&")
			records[i]["product_url"] = url
		}

		// Fix image_url
		if url, ok := records[i]["image_url"]; ok && url != "" {
			// Fix multiple slashes
			url = strings.Replace(url, "///thumbnail", "//thumbnail", -1)
			url = strings.Replace(url, "//thumbnail", "/thumbnail", -1)
			// Make sure we have proper protocol slashes
			if strings.HasPrefix(url, "https:/") && !strings.HasPrefix(url, "https://") {
				url = strings.Replace(url, "https:/", "https://", 1)
			}
			records[i]["image_url"] = url
		}
	}

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// Use a custom encoder that doesn't escape HTML characters
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	return encoder.Encode(records)
}

// saveAsCSV saves the data as CSV
func saveAsCSV(records []DataRecord, fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Extract headers from the first record
	if len(records) == 0 {
		return nil
	}

	var headers []string
	for key := range records[0] {
		headers = append(headers, key)
	}

	// Write headers
	if err := writer.Write(headers); err != nil {
		return err
	}

	// Write data
	for _, record := range records {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = record[header]
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// printSummary prints a summary of the scraped data
func printSummary(records []DataRecord) {
	fmt.Printf("\nScraping Summary:\n")
	fmt.Printf("Total records scraped: %d\n", len(records))

	// Print field distribution if records exist
	if len(records) > 0 {
		fmt.Println("\nField distribution:")
		for field := range records[0] {
			emptyCount := 0
			for _, record := range records {
				if record[field] == "" {
					emptyCount++
				}
			}
			fmt.Printf("  %s: %d/%d records have values (%.1f%% complete)\n",
				field, len(records)-emptyCount, len(records), float64(len(records)-emptyCount)*100/float64(len(records)))
		}
	}

	fmt.Println("\nData quality issues:")
	if len(records) == 0 {
		fmt.Println("  No records found. Check your selectors and URL.")
	} else {
		incompleteRecords := 0
		for _, record := range records {
			hasEmpty := false
			for _, value := range record {
				if value == "" {
					hasEmpty = true
					break
				}
			}
			if hasEmpty {
				incompleteRecords++
			}
		}
		fmt.Printf("  %d/%d records have missing values (%.1f%%)\n",
			incompleteRecords, len(records), float64(incompleteRecords)*100/float64(len(records)))
	}
}
