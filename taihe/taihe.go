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
	artistCollector := c.Clone()
	songCollector := c.Clone()
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
			artistCollector.Visit(e.Request.AbsoluteURL(link))
		}
	})

	// On every a element which has href attribute call callback
	artistCollector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if regSong.MatchString(link) {
			fmt.Printf("Song found: %q -> %s\n", strings.TrimSpace(strings.Replace(e.Text, "\n", "", -1)), link)
			// 歌曲列表
			songCollector.Visit(e.Request.AbsoluteURL(link))
		}
	})

	// 歌手对应的歌曲分页逻辑
	artistCollector.OnHTML("div.page-cont", func(e *colly.HTMLElement) {
		current := e.ChildText("span.page-navigator-current")
		next := e.ChildAttr("a.page-navigator-next", "href")
		fmt.Printf("Page song found: %q -> %s\n", current, next)
		// 歌曲列表 TODO
		// Page song found: "1" -> /data/style/getsongs?title=&start=15&size=15&third_type=
		// Page song found: "1" -> /data/style/getalbums?title=&start=12&size=12&third_type=
		// Page song found: "1" -> /data/artist/getmvlist?start=9&size=9&third_type=
		// Page song found: "" ->
		// c.Visit(e.Request.AbsoluteURL(link))
	})

	// On every ul element which has top_subnav__link class call callback
	songCollector.OnHTML("div.song-info-box", func(e *colly.HTMLElement) {
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
	c.Visit("http://music.taihe.com/artist/9103")

	outfile, _ := os.Create("taihe.json")
	enc := json.NewEncoder(outfile)
	enc.SetIndent("", "  ")
	// Dump json to the standard output
	enc.Encode(items)
}
