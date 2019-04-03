package main

import (
	"fmt"
	"time"

	"github.com/gocolly/colly"
	"github.com/wlcn/yq-colly/common"
	"github.com/wlcn/yq-colly/producer"
	"github.com/wlcn/yq-starter/service/article"
)

func main() {
	// Instantiate default collector
	c := colly.NewCollector()
	detailCollector := c.Clone()

	p := producer.NewSyncProducer()
	defer producer.CloseSync(p)

	c.OnHTML("div[id=articles]", func(e *colly.HTMLElement) {
		e.ForEach("article", func(_ int, a *colly.HTMLElement) {
			articleLink := a.ChildAttr("a.article-link", "href")
			excerpt := a.ChildText("section.post-excerpt p:first-child")
			// fmt.Printf("link is %v, excerpt is %v \n", articleLink, excerpt)
			e.Request.Ctx.Put("excerpt", excerpt)
			detailCollector.Request("GET", e.Request.AbsoluteURL(articleLink), nil, e.Request.Ctx, nil)
		})
		older := e.ChildAttr("nav.pagination a.older-posts", "href")
		fmt.Println(older)
		c.Visit(e.Request.AbsoluteURL(older))
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	detailCollector.OnHTML("div[id=article]", func(e *colly.HTMLElement) {
		publishTime, err := time.Parse("2006-01-02", e.ChildAttr("time.post-date", "datetime"))
		if err != nil {
			fmt.Printf("parse time err %+v \n", err)
			publishTime = time.Now()
		}
		content, err := e.DOM.Html()
		if err != nil {
			fmt.Printf("get detail error %+v", err)
			return
		}
		data := article.Article{
			Title:       e.ChildText("h1.post-title"),
			PublishTime: publishTime,
			Author:      e.ChildText("section.author a"),
			Content:     content,
		}
		producer.SendSync(p, common.Topic, data)
	})

	c.Visit("https://golangbot.com/")

}
