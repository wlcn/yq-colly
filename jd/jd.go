package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/wlcn/yq-colly/common"
)

func main() {
	c := colly.NewCollector()
	detailCollector := c.Clone()
	items := make([]common.Item, 0, 10)
	c.OnRequest(func(r *colly.Request) {
		log.Println("visiting ", r.URL.String())
	})

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*search.jd.com*",
		Parallelism: 2,
		RandomDelay: 10 * time.Second,
	})

	c.OnHTML(`a[href]`, func(e *colly.HTMLElement) {
		itemURL := e.Request.AbsoluteURL(e.Attr("href"))
		// 京东详情链接
		if strings.Index(itemURL, "item.jd.com") != -1 {
			detailCollector.Visit(itemURL)
		}
	})
	detailCollector.OnHTML(`body`, func(e *colly.HTMLElement) {
		log.Println("Item found", e.Request.URL)
		title := e.ChildText("div.sku-name")
		if title == "" {
			log.Print("No title found", e.Request.URL)
		} else {
			item := common.Item{
				Title:     title,
				SourceURL: e.Request.URL.String(),
				// Price:       e.ChildText("span.price"),
				Brand:       e.ChildText("ul[id=parameter-brand] li a"),
				Description: make(map[string]string, 10),
			}
			e.ForEach("ul.parameter2 li", func(_ int, el *colly.HTMLElement) {
				kv := strings.Split(el.Text, "：")
				item.Description[kv[0]] = kv[1]
			})
			e.ForEach("div.Ptable-item", func(_ int, el *colly.HTMLElement) {
				el.ForEach("dl.clearfix", func(_ int, el *colly.HTMLElement) {
					dt := el.ChildText("dt")
					dd := el.DOM.ChildrenFiltered("dd").Last().Text()
					item.Description[dt] = dd
				})
			})
			items = append(items, item)
		}
	})
	/*
		京东每页数据4*15=60
		100页数据应该为60*100=6000条
		但是京东显示的是7500+
	*/
	goods := []string{"洗衣机"}
	for _, good := range goods {
		key := url.QueryEscape(good)
		// 搜索产品按照新品排序
		for p := 1; p <= 100; p++ {
			s := 61 + 60*(p-2)
			log.Println(fmt.Sprintf("current page is %v", p))
			c.Visit(fmt.Sprintf("https://search.jd.com/Search?callback=jQuery1519158&area=1&enc=utf-8&keyword=%v&adType=7&page=%v&ad_ids=576%3A1&ad_type=4&_=1552566044492", key, strconv.Itoa(p*2-1)))
			c.Visit(fmt.Sprintf("https://search.jd.com/Search?keyword=%v&enc=utf-8&qrst=1&rt=1&stop=1&vt=2&wq=%v&psort=5&stock=1&page=%s&s=%s&click=0", key, key, strconv.Itoa(p*2-1), strconv.Itoa(s)))
		}
	}

	outfile, _ := os.Create("jd.json")
	enc := json.NewEncoder(outfile)
	enc.SetIndent("", "  ")
	// Dump json to the standard output
	enc.Encode(items)
}
