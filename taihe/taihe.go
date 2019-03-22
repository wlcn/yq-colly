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
	// "github.com/gocolly/colly/debug"
)

const urlTemplate = "http://music.taihe.com/data/user/getsongs?start=%d&size=%d&ting_uid=%s&.r=%.10f"
const baseURL = "http://music.taihe.com%v"
const urlTing = "http://musicapi.taihe.com/v1/restserver/ting?method=baidu.ting.song.playAAC&format=jsonp&callback=jQuery17203475326803441232_%v&songid=%v&from=web&_=%v"

// StoreDir 存储地址，目前歌曲，歌词，以及图片都存在一个目录，以歌曲[songId_songName_artist]命名
const StoreDir = "/tmp/taihe/"

var regArtist, _ = regexp.Compile(`^/artist/[\d]+$`)
var regSong, _ = regexp.Compile(`^/song/[\d]+$`)

func init() {
	rand.Seed(int64(time.Now().Nanosecond()))
}

// SongDetail struct
type SongDetail struct {
	ArtistID    string `json:"artistId"`
	Artist      string `json:"artist"`
	AlbumID     string `json:"albumId"`
	Album       string `json:"album"`
	SongID      string `json:"songId"`
	SongName    string `json:"songName"`
	SongLink    string `json:"songLink"`
	LrcLink     string `json:"lrcLink"`
	Publish     string `json:"publish"`
	Company     string `json:"company"`
	MVID        string `json:"mvId"`
	SongImgLink string `json:"songImgLink"`
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
		Lrclink string `json:"lrclink"`
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
	lrcCollector := c.Clone()
	audioCollector := c.Clone()

	items := make([]SongDetail, 0, 16)

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
		songInfo := e.DOM.Find("div.song-info-box")
		artistID, _ := songInfo.Find("span.artist a").Attr("href")
		albumID, _ := songInfo.Find("p.album a").Attr("href")
		songDetail := SongDetail{
			ArtistID:    e.Request.AbsoluteURL(artistID),
			Artist:      songInfo.Find("span.artist a").Text(),
			AlbumID:     e.Request.AbsoluteURL(albumID),
			Album:       songInfo.Find("p.album a").Text(),
			SongID:      e.Attr("data-songid"),
			SongName:    songInfo.Find("span.name").Text(),
			SongLink:    e.Request.URL.String(),
			LrcLink:     e.ChildAttr("div[id=lyricCont]", "data-lrclink"),
			Publish:     songInfo.Find("p.publish").Text(),
			Company:     songInfo.Find("p.company").Text(),
			MVID:        e.ChildAttr("div.song-img", "data-mvid"),
			SongImgLink: e.ChildAttr("img.music-song-ing", "src"),
		}
		items = append(items, songDetail)
		// log.Printf("Song Detail found: %+v\n", songDetail)
		e.Request.Ctx.Put("songInfo", fmt.Sprintf("%v_%v_%v", songDetail.SongID, songDetail.SongName, songDetail.Artist))
		// 请求歌曲信息
		songDetailLink := fmt.Sprintf(urlTing, time.Now().UnixNano()/1e6, songDetail.SongID, time.Now().UnixNano()/1e6)
		log.Printf("Song deatil link is %v", songDetailLink)
		songDetailCollector.Request("GET", songDetailLink, nil, e.Request.Ctx, nil)
	})

	// 保存歌词
	lrcCollector.OnResponse(func(r *colly.Response) {
		songInfo := r.Ctx.Get("songInfo")
		err := r.Save(StoreDir + songInfo + ".lrc")
		if err != nil {
			log.Print(err)
			log.Printf("保存歌词失败 %v", r.Request.URL.String())
		}
	})

	// 保存音频文件
	audioCollector.OnResponse(func(r *colly.Response) {
		// log.Println("下载音频文件中......")
		songInfo := r.Ctx.Get("songInfo")
		fileExtension := r.Ctx.Get("FileExtension")
		err := r.Save(StoreDir + songInfo + "." + fileExtension)
		if err != nil {
			log.Print(err)
			log.Printf("保存音频失败 %v", r.Request.URL.String())
		}
	})

	// 保存歌曲信息
	songDetailCollector.OnResponse(func(r *colly.Response) {
		songInfo := r.Ctx.Get("songInfo")
		bodyStr := string(r.Body)
		// 截取字符串
		jsonStr := bodyStr[strings.Index(bodyStr, "(")+1 : strings.LastIndex(bodyStr, ")")]
		// 保存json数据
		outfile, _ := os.Create(StoreDir + songInfo + ".json")
		enc := json.NewEncoder(outfile)
		enc.SetIndent("", "  ")
		enc.Encode(jsonStr)
		// log.Println(jsonStr)
		// 解析json
		var songData SongData
		json.Unmarshal([]byte(jsonStr), &songData)
		// log.Printf("songData is %+v", songData)
		// 下载lrc歌词
		lrcCollector.Request("GET", songData.SongInfo.Lrclink, nil, r.Ctx, nil)
		// 下载mp3
		r.Ctx.Put("FileExtension", songData.Bitrate.FileExtension)
		audioCollector.Request("GET", songData.Bitrate.FileLink, nil, r.Ctx, nil)
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
	c.Visit("http://music.taihe.com/artist/9809")

	outfileDone, _ := os.Create("taihe_done.json")
	encDone := json.NewEncoder(outfileDone)
	encDone.SetIndent("", "  ")
	// Dump json to the standard output
	encDone.Encode(items)
}
