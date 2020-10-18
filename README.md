# go-scraper

A simple web scraper written in Golang.

The scraper returns details about the supplied URL including information on the number and type of headings, and information on links found within the page.

## Running

The URL to scrape is supplied as an argument to the program:

```
go run main.go <URL>
```

## Testing

This project includes tests, they can be run with:

```
go test
```