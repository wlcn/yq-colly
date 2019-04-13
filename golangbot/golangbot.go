package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

func main() {
	// Instantiate default collector
	c := colly.NewCollector()
	detailCollector := c.Clone()
	ch := make(chan map[string]interface{}, 10)
	var wg sync.WaitGroup
	go save(ch, &wg)

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
		article := map[string]interface{}{
			"Title":       e.ChildText("h1.post-title"),
			"Excerpt":     e.Request.Ctx.Get("excerpt"),
			"PublishTime": publishTime,
			"Author":      e.ChildText("section.author a"),
			"Content":     content,
		}
		ch <- article
	})

	c.Visit("https://golangbot.com/")

	// 主线程等待goroutine结束
	wg.Wait()
}

func save(ch chan map[string]interface{}, wg *sync.WaitGroup) {
	for d := range ch {
		wg.Add(1)
		go send(d, wg)
	}
}

func send(d map[string]interface{}, wg *sync.WaitGroup) {
	url := "http://localhost:8085/api/v1/article"
	// 保存数据接口
	jsonStr, err := json.Marshal(d)
	if err != nil {
		fmt.Printf("json marshal err %v", err)
		return
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("token", "I am a valid token in YQ")
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("err is %+v \n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("response Status: %v \n", resp.Status)
	wg.Done()
}
