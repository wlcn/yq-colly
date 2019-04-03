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
	"github.com/wlcn/yq-colly/common"
	"github.com/wlcn/yq-colly/producer"
	"github.com/wlcn/yq-starter/service/article"
	// "github.com/gocolly/colly/debug"
)

const urlTemplate = "http://music.taihe.com/data/user/getsongs?start=%d&size=%d&ting_uid=%s&.r=%.10f"
const baseURL = "http://music.taihe.com%v"
const urlTing = "http://musicapi.taihe.com/v1/restserver/ting?method=baidu.ting.song.playAAC&format=jsonp&callback=jQuery17203475326803441232_%v&songid=%v&from=web&_=%v"

// StoreDir 存储地址，目前歌曲，歌词，以及图片都存在一个目录，以歌曲[songId_songName_artist]命名
const StoreDir = "/tmp/music/taihe/"

var regArtist, _ = regexp.Compile(`^/artist/[\d]+$`)
var regSong, _ = regexp.Compile(`^/song/[\d]+$`)

func init() {
	rand.Seed(int64(time.Now().Nanosecond()))
}

// PageData struct
type PageData struct {
	ErrorCode string `json:"errorCode"`
	Query     struct {
		Start   int     `json:"start"`
		Size    int     `json:"size"`
		TingUID int     `json:"ting_uid"`
		R       float64 `json:"_r"`
	} `json:"query"`
	Data struct {
		HTML string `json:"html"`
		Js   string `json:"js"`
		CSS  string `json:"css"`
	} `json:"data"`
}

// SongData struct
type SongData struct {
	ErrorCode int `json:"error_code"`
	SongInfo  struct {
		Lrclink    string `json:"lrclink"`
		Artist     string `json:"artist"`
		SongID     string `json:"song_id"`
		Title      string `json:"title"`
		Language   string `json:"language"`
		Country    string `json:"country"`
		Author     string `json:"author"`
		PicRadio   string `json:"pic_radio"`
		PicPremium string `json:"pic_premium"`
		PicSmall   string `json:"pic_small"`
		AlbumTitle string `json:"album_title"`
	} `json:"songinfo"`
	Bitrate struct {
		FileLink      string `json:"file_link"`
		FileSize      string `json:"file_size"`
		FileFormat    string `json:"file_format"`
		FileExtension string `json:"file_extension"`
	} `json:"bitrate"`
}

func init() {
	err := os.MkdirAll(StoreDir, os.ModePerm)
	if err != nil {
		log.Fatalf("创建 StoreDir 失败 %v", StoreDir)
	}
}

func main() {
	// Instantiate default collector
	c := colly.NewCollector(
		// Visit only domains: music.taihe.com, qukufile2.qianqian.com, musicapi.taihe.com, hangmenshiting.qianqian.com, zhangmenshiting.qianqian.com
		colly.AllowedDomains("music.taihe.com", "qukufile2.qianqian.com", "musicapi.taihe.com", "hangmenshiting.qianqian.com", "zhangmenshiting.qianqian.com"),
		// colly.Debugger(&debug.LogDebugger{}),
	)
	artistCollector := c.Clone()
	artistPageCollector := c.Clone()
	songCollector := c.Clone()
	songDetailCollector := c.Clone()

	p := producer.NewSyncProducer()
	defer producer.CloseSync(p)

	// Limit the number of threads started by colly to two
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*music.taihe.com*",
		Parallelism: 2,
		RandomDelay: 5 * time.Second,
	})
	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if regArtist.MatchString(link) {
			log.Printf("Artist found: %q -> %s\n", strings.TrimSpace(strings.Replace(e.Text, "\n", "", -1)), link)
			// 歌手列表
			artistCollector.Visit(e.Request.AbsoluteURL(link))
		}
	})

	// On every a element which has href attribute call callback
	artistCollector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if regSong.MatchString(link) {
			// log.Printf("Song found: %q -> %s\n", strings.TrimSpace(strings.Replace(e.Text, "\n", "", -1)), link)
			// 歌曲列表
			songCollector.Visit(e.Request.AbsoluteURL(link))
		}
	})

	// 歌手对应的歌曲分页逻辑
	artistCollector.OnHTML("div.song-list-box div.page-cont", func(e *colly.HTMLElement) {
		// log.Println(e.DOM.Html())
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

		start, okStart := m["start"]
		size, okSize := m["size"]
		if okStart && okSize && len(start) > 0 && len(size) > 0 {
			log.Printf("tingid is %v, start is %v, size is %v, songTotal is %v,", tingID, m["start"][0], m["size"][0], songTotal)
			sizeInt, _ := strconv.Atoi(m["size"][0])
			currInt, _ := strconv.Atoi(current)
			songTotalInt, _ := strconv.Atoi(songTotal)
			for i := currInt; i < songTotalInt; i++ {
				startInt := i * sizeInt
				link := fmt.Sprintf(urlTemplate, startInt, sizeInt, tingID, rand.Float64())
				log.Println("Page song found ", link)
				artistPageCollector.Visit(e.Request.AbsoluteURL(link))
			}
		}
	})

	// Verify response content
	artistPageCollector.OnResponse(func(r *colly.Response) {
		// log.Println("response", string(r.Body))
		var pageData PageData
		json.Unmarshal(r.Body, &pageData)
		// log.Println(pageData.Data.HTML)
		// Load the HTML document
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(pageData.Data.HTML))
		if err != nil {
			log.Println(err)
			return
		}
		// log.Println(doc.Html())
		// Find the review items
		doc.Find("a").Each(func(_ int, s *goquery.Selection) {
			// For each item found, get the band and title
			link, _ := s.Attr("href")
			if regSong.MatchString(link) {
				// title, _ := s.Attr("title")
				// log.Printf("Song found: %s -> %s\n", link, title)
				songCollector.Visit(fmt.Sprintf(baseURL, link))
			}
		})
	})

	// On every ul element which has top_subnav__link class call callback
	songCollector.OnHTML("div.songn-info-box", func(e *colly.HTMLElement) {
		SongID := e.Attr("data-songid")
		// 请求歌曲信息
		songDetailLink := fmt.Sprintf(urlTing, time.Now().UnixNano()/1e6, SongID, time.Now().UnixNano()/1e6)
		log.Printf("Song deatil link is %v", songDetailLink)
		e.Request.Ctx.Put("songDetailLink", songDetailLink)
		songDetailCollector.Request("GET", songDetailLink, nil, e.Request.Ctx, nil)
	})

	// 保存歌曲信息
	songDetailCollector.OnResponse(func(r *colly.Response) {
		songDetailLink := r.Ctx.Get("songDetailLink")
		bodyStr := string(r.Body)
		// 截取字符串
		jsonStr := bodyStr[strings.Index(bodyStr, "(")+1 : strings.LastIndex(bodyStr, ")")]
		// log.Println(jsonStr)
		// 解析json
		var songData SongData
		json.Unmarshal([]byte(jsonStr), &songData)
		// send to kafka
		data := article.Article{
			SourceID:    songData.SongInfo.SongID,
			URL:         songDetailLink,
			Title:       songData.SongInfo.Title,
			Content:     songData.SongInfo.AlbumTitle,
			PublishTime: time.Now(),
			Author:      songData.SongInfo.Author,
			Source:      "taihe",
			Tag:         "music",
			Lrclink:     songData.SongInfo.Lrclink,
			PicLink:     songData.SongInfo.PicPremium,
			FileLink:    songData.Bitrate.FileLink,
		}
		producer.SendSync(p, common.Topic, data)
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting ", r.URL.String())
	})

	// Verify response content
	c.OnResponse(func(r *colly.Response) {
		// log.Println("response", string(r.Body))
	})

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		log.Println("Request URL:", r.Request.URL, "failed with response:", r, "\n Error:", err)
	})
	// Start scraping on http://music.taihe.com/artist
	c.Visit("http://music.taihe.com/artist")

}
