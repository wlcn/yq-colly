package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gocolly/colly"
)

func main() {
	// Instantiate default collector
	c := colly.NewCollector()

	ch := make(chan map[string]interface{}, 10)
	var wg sync.WaitGroup
	go save(ch, &wg)

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnHTML(`div.demo`, func(e *colly.HTMLElement) {
		e.ForEach(`div.item_t`, func(_ int, d *colly.HTMLElement) {
			data := map[string]interface{}{
				"Title":   d.ChildText("span.price"),
				"URL":     e.Request.AbsoluteURL(d.ChildAttr("img", "src")),
				"Content": d.ChildAttr("img", "alt"),
				"Author":  d.ChildText("span.price"),
				"Source":  "xiaohuar",
				"Tag":     "image",
			}
			ch <- data
		})
		e.ForEach(`div.page a`, func(_ int, a *colly.HTMLElement) {
			if a.Text == "下一页" {
				nextLink := a.Attr("href")
				c.Visit(e.Request.AbsoluteURL(nextLink))
			}
		})
	})

	c.Visit("http://www.xiaohuar.com/hua/")
	// 主线程等待goroutine结束
	wg.Wait()
}

func save(ch chan map[string]interface{}, wg *sync.WaitGroup) {
	for d := range ch {
		wg.Add(1)
		go send(d, wg)
	}
}

func send(data map[string]interface{}, wg *sync.WaitGroup) {
	url := "http://localhost:8080/api/v1/image"
	// 保存数据接口
	jsonStr, err := json.Marshal(data)
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
