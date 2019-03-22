package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

const urlTemplate = "http://music.taihe.com/data/user/getsongs?start=%d&size=%d&ting_uid=%s&.r=%.10f"
const baseURL = "http://music.taihe.com%v"

var regArtist, _ = regexp.Compile(`^/artist/[\d]+$`)
var regSong, _ = regexp.Compile(`^/song/[\d]+$`)

func init() {
	rand.Seed(int64(time.Now().Nanosecond()))
}

// SongDetail struct
type SongDetail struct {
	ArtistID string `json:"artistId"`
	Artist   string `json:"artist"`
	AlbumID  string `json:"albumId"`
	Album    string `json:"album"`
	SongID   string `json:"songId"`
	SongName string `json:"songName"`
	SongLink string `json:"songLink"`
	LrcLink  string `json:"lrcLink"`
	Publish  string `json:"publish"`
	Company  string `json:"company"`
	MVID     string `json:"mvId"`
	SongImg  string `json:"songImg"`
}

// PageData struct
type PageData struct {
	Query struct {
		Start   int     `json:"start"`
		Size    int     `json:"size"`
		TingUID int     `json:"ting_uid"`
		R       float64 `json:"_r"`
	} `json:"query"`
	ErrorCode string `json:"errorCode"`
	Data      struct {
		HTML string `json:"html"`
		Js   string `json:"js"`
		CSS  string `json:"css"`
	} `json:"data"`
}

func main() {
	// Instantiate default collector
	c := colly.NewCollector(
		// Visit only domains: music.taihe.com
		colly.AllowedDomains("music.taihe.com"),
	)
	artistCollector := c.Clone()
	artistPageCollector := c.Clone()
	songCollector := c.Clone()
	items := make([]SongDetail, 0, 10)

	outfile, _ := os.Create("taihe.json")
	enc := json.NewEncoder(outfile)
	enc.SetIndent("", "  ")

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
			// fmt.Printf("Song found: %q -> %s\n", strings.TrimSpace(strings.Replace(e.Text, "\n", "", -1)), link)
			// 歌曲列表
			songCollector.Visit(e.Request.AbsoluteURL(link))
		}
	})

	// 歌手对应的歌曲分页逻辑
	artistCollector.OnHTML("div.song-list-box div.page-cont", func(e *colly.HTMLElement) {
		// fmt.Println(e.DOM.Html())
		current := e.ChildText("span.page-navigator-current")
		next := e.ChildAttr("a.page-navigator-next", "href")
		songTotal := e.DOM.Find("a.page-navigator-number").Last().Text()
		// albumsTotal := e.DOM.Find("a[class=page-navigator-number]").Last().Text()
		// mvTotal := e.DOM.Find("a[class=page-navigator-number]").Last().Text()
		next = e.Request.AbsoluteURL(next)

		// 解析ting_id
		p := e.Request.URL.Path
		tingID := p[strings.LastIndex(p, "/")+1:]
		u, _ := url.Parse(next)
		m, _ := url.ParseQuery(u.RawQuery)
		fmt.Printf("tingid is %v, start is %v, size is %v, songTotal is %v,", tingID, m["start"][0], m["size"][0], songTotal)

		sizeInt, _ := strconv.Atoi(m["size"][0])
		currInt, _ := strconv.Atoi(current)
		songTotalInt, _ := strconv.Atoi(songTotal)
		for i := currInt; i < songTotalInt; i++ {
			startInt := i * sizeInt
			link := fmt.Sprintf(urlTemplate, startInt, sizeInt, tingID, rand.Float64())
			fmt.Println("Page song found ", link)
			artistPageCollector.Visit(e.Request.AbsoluteURL(link))
		}
	})

	// Verify response content
	artistPageCollector.OnResponse(func(r *colly.Response) {
		// fmt.Println("response", string(r.Body))
		var pageData PageData
		json.Unmarshal(r.Body, &pageData)
		// fmt.Println(pageData.Data.HTML)
		// Load the HTML document
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(pageData.Data.HTML))
		if err != nil {
			log.Println(err)
			return
		}
		// fmt.Println(doc.Html())
		// Find the review items
		doc.Find("a").Each(func(_ int, s *goquery.Selection) {
			// For each item found, get the band and title
			link, _ := s.Attr("href")
			if regSong.MatchString(link) {
				// title, _ := s.Attr("title")
				// fmt.Printf("Song found: %s -> %s\n", link, title)
				songCollector.Visit(fmt.Sprintf(baseURL, link))
			}
		})
	})

	// On every ul element which has top_subnav__link class call callback
	songCollector.OnHTML("div.songn-info-box", func(e *colly.HTMLElement) {
		songInfo := e.DOM.Find("div.song-info-box")
		artistID, _ := songInfo.Find("span.artist a").Attr("href")
		albumID, _ := songInfo.Find("p.album a").Attr("href")
		songDetail := SongDetail{
			ArtistID: e.Request.AbsoluteURL(artistID),
			Artist:   songInfo.Find("span.artist a").Text(),
			AlbumID:  e.Request.AbsoluteURL(albumID),
			Album:    songInfo.Find("p.album a").Text(),
			SongID:   e.Attr("data-songid"),
			SongName: songInfo.Find("span.name").Text(),
			SongLink: e.Request.URL.String(),
			LrcLink:  e.ChildAttr("div[id=lyricCont]", "data-lrclink"),
			Publish:  songInfo.Find("p.publish").Text(),
			Company:  songInfo.Find("p.company").Text(),
			MVID:     e.ChildAttr("div.song-img", "data-mvid"),
			SongImg:  e.ChildAttr("img.music-song-ing", "src"),
		}
		// Dump json to the standard output
		enc.Encode(songDetail)
		items = append(items, songDetail)
		// fmt.Printf("Song Detail found: %+v\n", songDetail)
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
	c.Visit("http://music.taihe.com/artist/9809")

	outfileDone, _ := os.Create("taihe_done.json")
	encDone := json.NewEncoder(outfileDone)
	encDone.SetIndent("", "  ")
	// Dump json to the standard output
	encDone.Encode(items)
}
