# Web Scraping Project Report

## Approach and Design

### Project Overview

This project implements a general-purpose web scraper in Go that extracts structured data from e-commerce platforms. The scraper handles pagination, JavaScript-rendered content, and saves the results in structured JSON format based on user-defined records. This particular implementation focuses on extracting product details including names, prices, ratings, and images from e-commerce platforms.

Reference Websites:
- Shopee: https://shopee.sg/Computers-Peripherals-cat.11013247

##### Investigation of Anti-Scraping Techniques

To better understand e-commerce platforms' anti-scraping measures, I conducted several experiments:

###### Experiment 1: Simple HTTP Request

```bash
curl -A "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36"
```

The above curl command resulted in "Access Denied" errors, indicating that many e-commerce sites likely have implemented some form of protection against simple HTTP requests.
I then tried a more comprehensive curl command that includes various headers to mimic a real browser request:

```bash
curl --compressed -A "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36" \
     -H "Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9" \
     -H "Accept-Language: en-US,en;q=0.9" \
     -H "Accept-Encoding: gzip, deflate, br" \
     -H "Upgrade-Insecure-Requests: 1" \
     -H "Sec-Fetch-Dest: document" \
     -H "Sec-Fetch-Mode: navigate" \
     -H "Sec-Fetch-Site: none" \
     -H "Sec-Fetch-User: ?1" \
     "https://shopee.sg/Computers-Peripherals-cat.11013247"
```

With these headers that simulate a real browser request, the curl command does return the page content, indicating there are certain anti-scraping measures on many e-commerce websites.

##### Experiment 2: Headless Browser

When attempting to scrape Shopee, it was observed that product listings and other content are loaded dynamically using JavaScript. Due to this dynamic content loading, traditional static scraping methods (like those using GoQuery alone) were insufficient as they could not execute JavaScript to render the full page content.

Therefore, a headless browser solution, specifically ChromeDP, became essential. These tools allow the scraper to:
- Fully render web pages, including those with complex JavaScript.
- Simulate user interactions such as scrolling to load lazy-loaded content.
- Wait for specific elements to appear on the page before attempting to extract data.

This approach ensures that all product data, even that loaded asynchronously, is accessible for scraping, making the scraper robust for e-commerce platforms like Shopee.

### Technical Approach

#### Architecture

The scraper follows a modular design with the following components:

1. **Configuration Management**: Command-line flags for easy customization of scraping parameters.
2. **Scraping Engine**:
   - Dynamic scraping using ChromeDP for JavaScript-heavy sites like Shopee.
   - Static scraping using GoQuery for server-side rendered content.
3. **Data Extraction**: CSS selector-based extraction of both text content and attributes
4. **Pagination Handling**: Automatic detection and navigation of pagination elements
5. **Data Storage**: Flexible output in JSON or CSV formats
6. **Analysis**: Built-in data quality assessment and reporting

#### Technology Stack

- **Go**: Core programming language
- **GoQuery**: HTML parsing and CSS selector-based data extraction
- **ChromeDP**: Headless browser automation for JavaScript rendering
- **Standard Library**: JSON and CSV encoding/decoding

#### Anti-Scraping Measures

The scraper implements several techniques to handle anti-scraping measures encountered on e-commerce platforms:

1. **Rate Limiting**: Configurable delays between requests to avoid triggering rate limits
2. **Headless Browser**: Full JavaScript rendering capabilities for sites that heavily rely on client-side rendering and detect scraping through JS execution.
3. **User-Agent Spoofing**: Using a realistic browser user-agent to avoid detection
4. **Content Loading**: Implementing scrolling and waiting for dynamic content to load

### Implementation Details

#### CSS Selector Strategy

The scraper uses a two-level CSS selector approach:

1. **Record Selectors**: Identify individual data records (e.g., product listings) on the page.
2. **Field Selectors**: Extract specific fields from each identified record.

Examples of selectors used for different e-commerce platforms can be found under the `input` directory.

#### Pagination Handling

The scraper handles pagination differently for each platform:
- **Shopee**: Uses the CSS selector `button.shopee-icon-button--right` for its pagination button. But due to Shopee's dynamic nature and requires the user to login, we can't rely on the next button to advance the page, thus the need for manually appending page query param.

#### Data Quality Assurance

The scraper includes built-in data quality analysis that reports:

1. Total records extracted
2. Completeness of each field (percentage of records with non-empty values)
3. Overall data quality issues (records with missing values)

## Sample Data Analysis

### Dataset Overview

For this project, I scraped product data from Shopee's search results. The dataset includes:

- **Shopee**: 100 product records (limited for testing purposes) with 6 fields per record: product_name, price, discount_rate, sold, image_url, and product_url.

### Data Quality

The scraper successfully extracted product records with high data quality:

- **Shopee Data Quality**:
  - **product_name**: 100% complete
  - **price**: 100% complete
  - **discount_rate**: 95% complete (some products don't have discount)
  - **image_url**: 100% complete
  - **product_url**: 100% complete

### Data Insights

#### Price Distribution

The computer peripherals (for mouse) on Shopee range from budget options to more premium products. Analysis of the dataset reveals:

- **Price Range**: $4.22 to $139.00
- **Average Price**: $31.26
- **Price Segments**:
  - Budget options (<$10): Primarily basic mice and accessories
  - Mid-range ($10-$50): Most products fall in this category, including wireless mice with various features
  - Premium ($50+): High-end ergonomic and gaming mice with advanced features

#### Discount Patterns

Discount analysis reveals significant price reductions across the dataset:

- **Discount Range**: 1% to 78% off original prices
- **Average Discount**: 36.08%
- **Discount Distribution**: Most products feature discounts between 20-50%, with several high-discount outliers (70%+)

#### Brand Analysis

The dataset shows clear brand dominance in the computer peripherals category:

- **Logitech**: 55% of products (55 items) - dominant market leader
- **UGREEN**: 8% of products (8 items)
- **Razer**: 8% of products (8 items)
- **HP**: 4% of products (4 items)
- **Others**: 25% of products (various brands and unbranded items)

This brand distribution indicates Logitech's strong market position in the computer peripherals category on Shopee Singapore.

## Challenges and Solutions

### Challenge 1: Dynamic Content Loading and JavaScript Rendering

**Problem**: Websites like Shopee heavily rely on JavaScript to dynamically load content, making traditional static HTTP requests insufficient.

**Solution**: Implemented a headless browser approach using ChromeDP that fully renders the page before extraction. This allows for the execution of JavaScript and the rendering of all dynamic content. Added scrolling functionality and explicit waits for elements to ensure all lazy-loaded content is visible and interactive elements are ready.

### Challenge 2: Rate Limit

**Problem**: Rate limiting on e-commerce platforms can lead to the scraping being blocked.

**Solution**: Implemented a wait time between requests to avoid triggering rate limits. But of course this is not a perfect solution, as it's not guaranteed that the rate limit will be lifted. A better way would probably be to use a proxy server to distribute requests across different IP addresses.

### Challenge 3: Stability

**Problem**: The crawler itself is not very stable and very susceptible to timing out

**Solution**: Implemented a simple retry mechanism, which will retry the request up to 3 times if it fails. This is not a perfect solution, and a better way should be investigating the root cause of instability. But I expect this will always be the cause and we should design the solution to be more fault-tolerant and robust.

## Conclusion

The implemented web scraper successfully meets all the project requirements:

1. **Data Extraction**: Extracts structured data from various e-commerce categories with robust support for JavaScript-rendered content.
2. **Pagination Handling**: Automatically navigates through multiple pages, adapting to different pagination mechanisms, to collect comprehensive datasets.
3. **Structured Output**: Saves extracted data in well-formatted JSON files, ensuring easy consumption and analysis.
4. **Anti-Scraping Measures**: Implements various techniques to avoid detection and blocking, enhancing the scraper's resilience.
5. **Data Analysis**: Provides insights into the quality and characteristics of the extracted data, aiding in understanding the scraped content.

The modular design ensures the scraper can be easily adapted to different categories on existing platforms or even new e-commerce websites with minimal configuration changes, demonstrating its versatility and extensibility.

## Future Improvements

Potential enhancements for future versions include:

1. **Proxy Support**: Implement rotation of proxy servers to distribute requests across different IP addresses
2. **Category Navigation**: Add support for automatically navigating through different product categories
3. **Product Detail Scraping**: Enhance the scraper to visit individual product pages and extract more detailed information
4. **Multi-platform Support**: Extend the scraper to handle more e-commerce platforms with different structures and anti-scraping measures
