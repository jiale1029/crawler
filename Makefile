shopee:
	go run main.go -max=100 -url="https://shopee.sg/Computers-Peripherals-cat.11013247" -record="li.shopee-search-item-result__item" --pagination="nav.shopee-page-controller a.shopee-icon-button.shopee-icon-button--right" --mapping input/shopee_fields.json --output output/shopee

shopee_search:
	go run main.go -max=100 -url="https://shopee.sg/search?keyword=mouse" -record="li.shopee-search-item-result__item" --pagination="nav.shopee-page-controller a.shopee-icon-button.shopee-icon-button--right" --mapping input/shopee_search_field.json --output output/shopee_search