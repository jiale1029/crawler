# Go Web Scraper

A flexible web scraper built in Go, designed to extract product information from e-commerce websites. The scraper can handle JavaScript-rendered content, pagination, and provides structured data output in JSON or CSV format.

## Features

- **Configurable CSS Selectors**: Easy customization for different websites and data fields
- **Pagination Support**: Automatically navigate through multiple pages
- **Structured Output**: Save data in JSON or CSV formats
- **Data Quality Analysis**: Built-in reporting on data completeness
- **Anti-Scraping Measures**: Configurable delays, user-agent spoofing, and JavaScript rendering

## Installation

### Installing Go

Before running the scraper, you need to have Go installed on your system. Follow these steps to install Go:

#### macOS

1. **Using Homebrew (recommended)**:
   ```bash
   brew install go
   ```

2. **Using the official installer**:
   - Download the latest Go package from [golang.org/dl](https://golang.org/dl/)
   - Follow the installation wizard
   - Verify installation with `go version`

#### Linux

1. **Using package manager**:
   ```bash
   # For Ubuntu/Debian
   sudo apt update
   sudo apt install golang-go
   
   # For CentOS/RHEL
   sudo yum install golang
   ```

2. **Using the official binary**:
   ```bash
   wget https://golang.org/dl/go1.21.0.linux-amd64.tar.gz
   sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
   ```
   Then add Go to your PATH in ~/.profile or ~/.bashrc:
   ```bash
   export PATH=$PATH:/usr/local/go/bin
   export GOPATH=$HOME/go
   ```

### Verify Go Installation

After installation, verify that Go is correctly installed by running:

```bash
go version
```

You should see output showing the installed Go version.

### Setting up the Project

```bash
# Clone the repository
git clone https://github.com/jiale1029/crawler.git
cd crawler

# Install dependencies
go mod tidy

# Run the scraper
make shopee
```

## Usage

### Quick Start with Make Commands

```bash
# For scraping Shopee computer peripherals
make shopee
```

### Manual Configuration

```bash
# Basic usage with default settings
./scraper

# Limit the number of records to extract
./scraper -max=50

# Specify output format and filename
./scraper -format=json -output="product_data.json"
```

## Configuration Options

The scraper supports the following command-line flags:

| Flag          | Description                           | Default                                     |
| ------------- | ------------------------------------- | ------------------------------------------- |
| `-url`        | URL to scrape                         | `https://www.example.com/products/category` |
| `-mapping`    | Mapping of the field and selector     | `input/fields.json`                           |
| `-output`     | Output filename                       | `output.json`                                |
| `-format`     | Output format (json or csv)           | `json`                                      |
| `-max`        | Maximum number of records to extract  | `100`                                       |
| `-wait`       | Wait time between requests in seconds | `2`                                         |
| `-timeout`    | Timeout configuration for request      | `45`                                        |
| `-pagination` | CSS selector for pagination element   | `a.btn-next`                                |
| `-record`     | CSS selector for record elements      | `.product-item`                             |

## Customizing Field Extraction

### Shopee Fields

For Shopee, the scraper extracts these fields:

- `product_name`: Product title
- `price`: Current price
- `discount_rate`: Discount percentage
- `sold`: Number of items sold
- `image_url`: Product image URL
- `product_url`: Link to product page

### Custom Field Configuration

To customize field extraction, modify the corresponding JSON files in the `input` directory or create a new one:

```json
{
    "product_name": ".css-selector-for-name",
    "price": ".css-selector-for-price",
    "custom_field": ".css-selector-for-custom-field"
}
```

### Attribute Extraction

To extract attributes from HTML elements, use the `@attr:` syntax in your CSS selector:

```json
{
    "image_url": ".image img@attr:src",
    "product_url": "a.product-link@attr:href"
}
```

## Handling Anti-Scraping Measures

The scraper includes several features to handle anti-scraping measures:

1. **Rate Limiting**: Use the `-wait` flag to add delays between requests
2. **User-Agent Spoofing**: The scraper uses a realistic browser user-agent
3. **JavaScript Rendering**: Dynamic scraping with ChromeDP handles JavaScript-rendered content
4. **Content Loading**: Implements scrolling to ensure all lazy-loaded content is visible

## Troubleshooting

### Common Issues

1. **Code Hanging**: If the scraper hangs at the browser automation section, it might be waiting for elements that don't appear. Check that your CSS selectors for `-record` and `-pagination` are correct for the website you're scraping. For example:
   - For Shopee: `-record=".shopee-search-item-result__item" -pagination="button.shopee-icon-button--right"`

2. **No Data Extracted**: Verify that the CSS selectors in your fields configuration match the actual HTML structure of the website. It could be timeout issue as well, check the logs and increase the timeout if needed.

3. **Pagination Not Working**: Make sure the pagination selector correctly identifies the "next page" element. Different websites use different HTML structures for pagination.


## Data Quality Analysis

After scraping, the tool provides a summary of data quality:

```
Scraping completed. Extracted 20 records.
Data quality summary:
- product_name: 100.00% complete
- price: 100.00% complete
- original_price: 85.00% complete
- discount_rate: 0.00% complete
- rating: 100.00% complete
- review_count: 100.00% complete
- image_url: 100.00% complete
- product_url: 100.00% complete

Records with missing values: 20 (100.00%)
```

## License

MIT
