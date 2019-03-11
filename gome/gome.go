package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

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
	c.OnHTML(`span.min-pager-number`, func(e *colly.HTMLElement) {
		// 解析下一页
		fl := e.Text
		page := strings.TrimSpace(fl[0:strings.Index(fl, "/")])
		total := strings.TrimSpace(fl[strings.Index(fl, "/")+1:])
		intpage, _ := strconv.Atoi(page)
		inttotal, _ := strconv.Atoi(total)
		if intpage >= inttotal {
			log.Println("download page is done")
		} else {
			// 页数加一，计算下一个分页链接参数
			log.Println(fmt.Sprintf("total page is %v, current page is %v", total, page))
			itemURL := e.Request.URL.String()
			u, _ := url.Parse(itemURL)
			values, _ := url.ParseQuery(u.RawQuery)
			kw := url.QueryEscape(values["question"][0])
			intpage++
			c.Visit(fmt.Sprintf("https://search.gome.com.cn/search?question=%s&searchType=goods&search_mode=normal&reWrite=true&page=%s", kw, strconv.Itoa(intpage)))
		}
	})
	c.OnHTML(`a[href]`, func(e *colly.HTMLElement) {
		itemURL := e.Request.AbsoluteURL(e.Attr("href"))
		if strings.Index(itemURL, "item.gome.com.cn") != -1 {
			detailCollector.Visit(itemURL)
		}
	})

	// Extract details of the Item
	detailCollector.OnHTML(`ul.specbox`, func(e *colly.HTMLElement) {
		log.Println("Item found", e.Request.URL)
		item := common.Item{
			// Title:     title,
			SourceURL: e.Request.URL.String(),
			// Price:       e.ChildText("span.price"),
			// Brand:       e.ChildText("ul[id=parameter-brand] li a"),
			Description: make(map[string]string, 10),
		}
		e.ForEach("li", func(_ int, el *colly.HTMLElement) {
			name := el.ChildText("span.specinfo")
			val := el.DOM.ChildrenFiltered("span").Last().Text()
			if name != "" && val != "" {
				item.Description[name] = val
			}
		})
		items = append(items, item)
	})
	// 洗衣机编码格式
	goods := []string{"洗衣机"}
	for _, good := range goods {
		key := url.QueryEscape(good)
		c.Visit(fmt.Sprintf(`https://search.gome.com.cn/search?question=%s&searchType=goods&search_mode=normal&reWrite=true&page=1`, key))
	}

	outfile, _ := os.Create("gome.json")
	enc := json.NewEncoder(outfile)
	enc.SetIndent("", "  ")
	// Dump json to the standard output
	enc.Encode(items)
}
