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

// Article should only be concerned with database schema, more strict checking should be put in validator.
type Article struct {
	Title       string
	Content     string
	Author      string
	Excerpt     string
	ReadCount   int
	LikeCount   int
	PublishTime time.Time
	UserID      string
}

func main() {
	// Instantiate default collector
	c := colly.NewCollector()
	detailCollector := c.Clone()
	ch := make(chan Article, 10)
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
		article := Article{
			Title:       e.ChildText("h1.post-title"),
			Excerpt:     e.Request.Ctx.Get("excerpt"),
			PublishTime: publishTime,
			Author:      e.ChildText("section.author a"),
			Content:     content,
		}
		ch <- article
	})

	c.Visit("https://golangbot.com/")

	// 主线程等待goroutine结束
	wg.Wait()
}

func save(ch chan Article, wg *sync.WaitGroup) {
	for article := range ch {
		wg.Add(1)
		url := "http://localhost:8080/api/v1/article"
		// 保存数据接口
		token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoid2wiLCJwYXNzd29yZCI6IiQyYSQxMCRhRDJ0dmd2U09NYTRWMmxWdmh1LlRlcHlTOUhCZk9WOS5mdFdkbTlBWTVyUE5QczhKS1E5dSIsImV4cCI6MTU1Mzg1MTMxNiwiaXNzIjoieXEtc3RhcnRlci1pc3N1ZXIifQ._BR6lLyDBUt9PQwfc1LyfBVHKW8CyEdDzGWIKXIexa8"
		var jsonStr, err = json.Marshal(article)
		if err != nil {
			fmt.Printf("json marshal err %v", err)
			return
		}
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
		req.Header.Set("token", token)
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("err is %+v \n", err)
			return
		}
		defer resp.Body.Close()
		fmt.Printf("title is %v response Status: %v \n", article.Title, resp.Status)
		// fmt.Println("response Headers:", resp.Header)
		// body, _ := ioutil.ReadAll(resp.Body)
		// fmt.Println("response Body:", string(body))
		wg.Done()
	}
}
