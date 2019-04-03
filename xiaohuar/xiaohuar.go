package main

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
	"github.com/wlcn/yq-colly/common"
	"github.com/wlcn/yq-colly/producer"
	"github.com/wlcn/yq-starter/service/image"
)

func main() {
	// Instantiate default collector
	c := colly.NewCollector()

	p := producer.NewSyncProducer()
	defer producer.CloseSync(p)

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnHTML(`html`, func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		e.ForEach(`a`, func(_ int, a *colly.HTMLElement) {
			itemURL := e.Request.AbsoluteURL(e.Attr("href"))
			if strings.Index(itemURL, "www.xiaohuar.com") != -1 {
				c.Visit(itemURL)
			}
		})
		if strings.Index(url, "http://www.xiaohuar.com/list-") != -1 {
			e.ForEach("div.img a", func(_ int, a *colly.HTMLElement) {
				data := image.Image{
					Title:   a.Attr("alt"),
					URL:     a.Attr("href"),
					Content: a.Text,
				}
				producer.SendSync(p, common.Topic, data)
			})
		}
	})

	c.Visit("http://www.xiaohuar.com/hua/")

}
