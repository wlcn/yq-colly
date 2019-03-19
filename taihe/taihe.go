package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

const urlTemplate = "http://music.taihe.com/data/user/getsongs?start=%d&size=%d&ting_uid=%s&.r=%s"
const lrcURLTemplate = "http://music.taihe.com/data/song/lrc?lrc_link=%s"

var regArtist, _ = regexp.Compile(`^/artist/[\d]+$`)
var regSong, _ = regexp.Compile(`^/song/[\d]+$`)

func main() {
	// Instantiate default collector
	c := colly.NewCollector(
		// Visit only domains: music.taihe.com
		colly.AllowedDomains("music.taihe.com"),
	)
	detailCollector := c.Clone()
	items := make([]string, 0, 10)
	// Limit the number of threads started by colly to two
	// when visiting links which domains' matches "*httpbin.*" glob
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*music.taihe.com*",
		Parallelism: 2,
		RandomDelay: 5 * time.Second,
	})
	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if regArtist.MatchString(link) {
			fmt.Printf("Artist found: %q -> %s\n", strings.TrimSpace(strings.Replace(e.Text, "\n", "", -1)), link)
			// 歌手列表
			c.Visit(e.Request.AbsoluteURL(link))
		} else if regSong.MatchString(link) {
			fmt.Printf("Song found: %q -> %s\n", strings.TrimSpace(strings.Replace(e.Text, "\n", "", -1)), link)
			// 歌曲列表
			detailCollector.Visit(e.Request.AbsoluteURL(link))
		}
	})

	// 歌曲分页逻辑
	c.OnHTML("div.page-cont", func(e *colly.HTMLElement) {
		current := e.ChildText("span.page-navigator-current")
		next := e.ChildAttr("a.page-navigator-next", "href")
		fmt.Printf("Page song found: %q -> %s\n", current, next)
		// 歌曲列表 TODO
		// c.Visit(e.Request.AbsoluteURL(link))
	})

	// On every ul element which has top_subnav__link class call callback
	detailCollector.OnHTML("div.song-info-box", func(e *colly.HTMLElement) {
		songName := e.ChildText("span.name")
		artist := e.ChildText("span.artist a")
		fmt.Printf("Song detail found: %v, %q\n", songName, artist)
		
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
	// Start scraping on http://music.taihe.com/artist
	c.Visit("http://music.taihe.com/artist")

	outfile, _ := os.Create("taihe.json")
	enc := json.NewEncoder(outfile)
	enc.SetIndent("", "  ")
	// Dump json to the standard output
	enc.Encode(items)
}
