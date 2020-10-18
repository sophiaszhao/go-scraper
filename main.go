package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/olekukonko/tablewriter"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

//How long to wait between requests
const delayBetweenRequests = time.Second

//The different types of links
type linkDetails struct {
	nAnchor   int
	nInternal int
	nExternal int
	nInvalid  int
}

func main() {
	args := os.Args[1:]

	if len(args) != 1 {
		log.Fatal("Usage: go-scraper URL")
	}

	baseURL, err := url.Parse(args[0])
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}

	req, err := http.NewRequest("GET", baseURL.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	//If the URL does not return a 200 response then fail.
	if resp.StatusCode != 200 {
		log.Fatal("status code error:", resp.Status)
	}

	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	//Find HMTL version (if it's html5)
	version := "Unknown"
	if doc.FirstChild != nil && doc.FirstChild.Type == html.DoctypeNode {
		doctype := doc.FirstChild
		if doctype.Data == "html" {
			version = "HTML5"
		}
	}

	fmt.Println("HTML Version:", version)

	//Find title of the page
	title, err := htmlquery.Query(doc, "//title")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Title:", htmlquery.InnerText(title))

	//Find login, a form with element of <input type == "password">
	loginForms, err := htmlquery.QueryAll(doc, "//form[//input[@type='password']]")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Login page:", len(loginForms) > 0)

	//Calculate the number of headings of each type
	err = processHeadings(doc)
	if err != nil {
		log.Fatal(err)
	}

	//Calculate the number of links
	err = processLinks(client, baseURL, doc)
	if err != nil {
		log.Fatal(err)
	}
}

func processHeadings(doc *html.Node) error {
	headerCounts, err := countHeadings(doc)
	if err != nil {
		return err
	}

	fmt.Println("Heading details:")

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Level", "Quantity"})

	levels := []string{"h1", "h2", "h3", "h4", "h5", "h6"}
	for _, level := range levels {
		table.Append([]string{level, strconv.Itoa(headerCounts[level])})
	}

	table.Render()

	return nil
}

func countHeadings(doc *html.Node) (map[string]int, error) {
	headerCounts := make(map[string]int)
	headers, err := htmlquery.QueryAll(doc, "//*[self::h1 or self::h2 or self::h3 or self::h3 or self::h4 or self::h5 or self::h6]")
	if err != nil {
		return nil, err
	}
	for _, header := range headers {
		headerCounts[header.Data]++
	}

	return headerCounts, nil
}

func processLinks(client *http.Client, baseURL *url.URL, doc *html.Node) error {
	details, err := countLinks(client, baseURL, doc)
	if err != nil {
		return err
	}

	data := [][]string{
		[]string{"Anchor", strconv.Itoa(details.nAnchor)},
		[]string{"Internal", strconv.Itoa(details.nInternal)},
		[]string{"External", strconv.Itoa(details.nExternal)},
		[]string{"Inaccessible", strconv.Itoa(details.nInvalid)},
	}
	fmt.Println("Link details:")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Link Type", "Quantity"})

	for _, t := range data {
		table.Append(t)
	}
	table.Render()

	return nil
}

func countLinks(client *http.Client, baseURL *url.URL, doc *html.Node) (*linkDetails, error) {
	details := linkDetails{}

	nodes, err := htmlquery.QueryAll(doc, "//a")
	if err != nil {
		return nil, err
	}

	links := make([]*url.URL, 0)
	for _, node := range nodes {
		href := htmlquery.SelectAttr(node, "href")
		parsedLink, err := url.Parse(href)
		if err != nil {
			details.nInvalid++
			continue
		}

		resolvedHyperLink := baseURL.ResolveReference(parsedLink)
		resolvedHyperLink.Fragment = ""

		//Not an Anchor link which points to content on the same page
		if resolvedHyperLink.String() != baseURL.String() {
			links = append(links, resolvedHyperLink)
		} else {
			details.nAnchor++
		}
	}

	fmt.Println("Checking links for validity ...")

	//Check if the links are inaccessible
	err = validateLinks(client, baseURL, links, &details)
	if err != nil {
		return nil, err
	}

	return &details, nil
}

func validateLinks(client *http.Client, baseURL *url.URL, links []*url.URL, results *linkDetails) error {
	//Show a progress bar
	bar := pb.StartNew(len(links))

	for _, link := range links {
		if link.Hostname() == baseURL.Hostname() {
			results.nInternal++
		} else {
			results.nExternal++
		}

		//Find inaccessible link, use http head request, if not 200/err then invalid
		req, err := http.NewRequest("HEAD", link.String(), nil)
		if err != nil {
			return err
		}

		resp, err := client.Do(req)
		if err != nil {
			results.nInvalid++
			continue
		}

		err = resp.Body.Close()
		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			results.nInvalid++
		}

		bar.Increment()
		time.Sleep(delayBetweenRequests)
	}

	bar.Finish()

	return nil
}
