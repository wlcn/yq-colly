package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
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
	c.OnHTML(`div.product-box a`, func(e *colly.HTMLElement) {
		itemURL := e.Request.AbsoluteURL(e.Attr("href"))
		if strings.Index(itemURL, "product.suning.com") != -1 {
			detailCollector.Visit(itemURL)
		}
	})

	detailCollector.OnHTML(`table[id=itemParameter]`, func(e *colly.HTMLElement) {
		log.Println("Item found", e.Request.URL)
		item := common.Item{
			SourceURL: e.Request.URL.String(),
			// Price:       e.ChildText("span.price"),
			Description: make(map[string]string, 10),
		}
		e.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			name := el.ChildText("td.name")
			val := el.ChildText("td.val")
			if name != "" && val != "" {
				item.Description[name] = val
			}
		})
		items = append(items, item)
	})

	/*
		苏宁网站分析
		总43页
		每页加载完整是120条数据，分4次下拉加载
		120/4=30
		每次加载30个数据
		总数据应该是
		120*42=5040条数据
		最后一页只有6条数据
		目前抓取的数据有4654，相差400条左右，中间可能夹杂一些其他链接数据(不包含product.suning.com会过滤掉)导致过滤掉一些
		search.suning.com
		https://search.suning.com/emall/searchV1Product.do?keyword=%E6%B4%97%E8%A1%A3%E6%9C%BA&ci=244006&pg=01&cp=4&il=0&st=0&iy=0&adNumber=0&isDoufu=1&isNoResult=0&n=1&sesab=ACAABAAB&id=IDENTIFYING&cc=025&paging=2&sub=0&jzq=5073
	*/
	goods := []string{"洗衣机"}
	for _, good := range goods {
		key := url.QueryEscape(good)
		// cp=[0:43] paging[0:4] 前闭后开
		for cp := 0; cp < 43; cp++ {
			for p := 0; p < 4; p++ {
				if p == 0 {
					c.Visit(fmt.Sprintf(`https://search.suning.com/emall/searchV1Product.do?keyword=%v&ci=244006&pg=01&cp=%v&il=0&st=0&iy=0&adNumber=5&isDoufu=1&isNoResult=0&n=1&sesab=ACAABAAB&id=IDENTIFYING&cc=025&sub=0&jzq=5071`, key, cp))
				} else {
					c.Visit(fmt.Sprintf(`https://search.suning.com/emall/searchV1Product.do?keyword=%v&ci=244006&pg=01&cp=%v&il=0&st=0&iy=0&adNumber=5&isDoufu=1&isNoResult=0&n=1&sesab=ACAABAAB&id=IDENTIFYING&cc=025&paging=%v&sub=0&jzq=5071`, key, cp, p))
				}
			}
		}
	}

	outfile, _ := os.Create("suning.json")
	enc := json.NewEncoder(outfile)
	enc.SetIndent("", "  ")
	// Dump json to the standard output
	enc.Encode(items)
}
