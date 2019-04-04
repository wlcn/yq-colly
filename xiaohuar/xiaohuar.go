package main

import (
	"fmt"

	"github.com/gocolly/colly"
	"github.com/wlcn/yq-colly/common"
	"github.com/wlcn/yq-colly/producer"
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

	c.OnHTML(`div.demo`, func(e *colly.HTMLElement) {
		e.ForEach(`div.item_t`, func(_ int, d *colly.HTMLElement) {
			data := map[string]string{
				"Title":   d.ChildText("span.price"),
				"URL":     e.Request.AbsoluteURL(d.ChildAttr("img", "src")),
				"Content": d.ChildAttr("img", "alt"),
				"Author":  d.ChildText("span.price"),
				"Source":  "xiaohuar",
				"Tag":     "image",
			}
			producer.SendSync(p, common.TopicImage, data)
		})
		e.ForEach(`div.page a`, func(_ int, a *colly.HTMLElement) {
			if a.Text == "下一页" {
				nextLink := a.Attr("href")
				c.Visit(e.Request.AbsoluteURL(nextLink))
			}
		})
	})

	c.Visit("http://www.xiaohuar.com/hua/")

}
