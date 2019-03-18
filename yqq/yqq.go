package main

import (
	"fmt"

	"github.com/gocolly/colly"
)

func main() {
	// Instantiate default collector
	c := colly.NewCollector(
		// Visit only domains: y.qq.com, qq.com
		colly.AllowedDomains("y.qq.com", "qq.com"),
	)

	// On every a element which has top_subnav__link class call callback
	c.OnHTML("a.top_subnav__link", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// Print link
		fmt.Printf("Link found: %q -> %s\n", e.Text, link)
		// Visit link found on page
		// Only those links are visited which are in AllowedDomains
		c.Visit(e.Request.AbsoluteURL(link))
	})

	// On every ul element which has top_subnav__link class call callback
	c.OnHTML("ul.songlist__list", func(e *colly.HTMLElement) {
		e.ForEach("li.songlist__item", func(_ int, li *colly.HTMLElement) {
			// fmt.Println(li.DOM.Html())
			song := li.ChildText("h3.songlist__song a")
			singerName := li.ChildText("p.songlist__author a.singer_name")
			itemURL := li.ChildAttr("img.songlist__pic", "src")
			fmt.Printf("Item found: %v, %q -> %s\n", song, singerName, li.Request.AbsoluteURL(itemURL))
			// Visit link found on page
			// Only those links are visited which are in AllowedDomains
			// c.Visit(e.Request.AbsoluteURL(link))
		})
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting ", r.URL.String())
	})

	// Verify response content
	c.OnResponse(func(r *colly.Response) {
		// fmt.Println("response", string(r.Body))
	})

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\n Error:", err)
	})
	// Start scraping on https://y.qq.com/
	c.Visit("https://y.qq.com/")
}
