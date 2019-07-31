package novelrss

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/djimenez/iconv-go"
	"github.com/gorilla/feeds"
	"github.com/pkg/errors"
)

type Chapter struct {
	Author string    `json:"author"`
	Title  string    `json:"title"`
	Link   string    `json:"link"`
	MD5    string    `json:"md5"`
	Time   time.Time `json:"updated_at"`
	IsNew  bool
}

type NovelItem struct {
	Title         string `json:"title"`
	Url           string `json:"url"`
	UpdateChapter Chapter
}

type NovelList []NovelItem

func (l NovelList) Len() int {
	return len(l)
}

func (l NovelList) Less(i, j int) bool {
	return l[i].UpdateChapter.Time.After(l[j].UpdateChapter.Time)
}

func (l NovelList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

type NovelSite struct {
	Selector map[string]string
	BaseUrl  string
}

var UpdateTime time.Time

var selectorList = map[string]NovelSite{
	"uukanshu": {
		Selector: map[string]string{
			"newchapter": ".zuixin",
			"author":     ".jieshao_content",
		},
		BaseUrl: "https://www.uukanshu.com",
	},
	//	"qidian":   {Selector: ".update", BaseUrl: "https:"},
	//	"shuqu8":   {Selector: ".stats .fl", BaseUrl: "http://www.shuqu8.com/"},
}

func getUpdateChapter(novelUrl string) (Chapter, error) {

	var update Chapter
	var err error = nil

	defer func() {
		if err != nil {
			log.Println(err)
		}
	}()

	resp, err := http.Get(novelUrl)
	if err != nil {
		return update, errors.Wrap(err, "http request failed")
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return update, errors.Wrap(err, "http response failed")
	}

	md5 := md5.Sum([]byte(body))
	update.MD5 = hex.EncodeToString(md5[:16])
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return update, errors.Wrap(err, "goquery create document failed")
	}

	selector := map[string]string{
		"newchapter": ".zuixin",
		"author":     ".jieshao_content",
	}
	update.Link = "http://"
	for k, v := range selectorList {
		if strings.Contains(novelUrl, k) {
			selector = v.Selector
			update.Link = v.BaseUrl
		}
	}
	update.Title = "not found"
	// Find the review items
	fmt.Printf("=== selecor (%s)===\n", selector["newchapter"])
	doc.Find(selector["newchapter"]).Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		title := s.Find("a").Text()
		link, _ := s.Find("a").Attr("href")
		if utf8.ValidString(title) == false {
			update.Title, err = iconv.ConvertString(title, "gb2312", "utf-8")
			if err != nil {
				log.Println(err.Error())
				update.Title = "not utf8"
			}
		} else {
			update.Title = title
		}
		if link[0] != '/' {
			update.Link = novelUrl + link
		} else {
			update.Link = update.Link + link
		}
		update.Time = time.Now()
		log.Printf("Scrap Novel Chapter %d: %s %s %s %s\n", i, update.Title, update.Link, update.Time.Format("2006-01-02 15:04:05"), update.MD5)
	})
	return update, nil
}

func Update(novellist NovelList) {

	for i, item := range novellist {
		update, err := getUpdateChapter(item.Url)
		if err != nil {
			log.Println(err)
		} else if novellist[i].UpdateChapter.MD5 != update.MD5 && novellist[i].UpdateChapter.Title != update.Title {
			novellist[i].UpdateChapter = update
			novellist[i].UpdateChapter.IsNew = true
			fmt.Printf("Update Novel Chapter: [ %s ] %s %s %s %s\n", item.Title, update.Title, update.Link, update.Time.Format("2006-01-02 15:04:05"), update.MD5)
		}
	}
	sort.Sort(novellist)
	fmt.Println(novellist)
	UpdateTime = time.Now()
	fmt.Printf("Update at:%s\n", time.Now().Format("2006-01-02 15:04:05"))

	for {
		select {
		case <-time.After(10 * time.Minute):
			for i, item := range novellist {
				update, err := getUpdateChapter(item.Url)
				if err != nil {
					log.Println(err)
				} else if novellist[i].UpdateChapter.MD5 != update.MD5 && novellist[i].UpdateChapter.Title != update.Title {
					old := novellist[i].UpdateChapter
					fmt.Printf("Origin Novel Chapter: [ %s ] %s %s %s %s\n", item.Title, old.Title, old.Link, old.Time.Format("2006-01-02 15:04:05"), old.MD5)
					novellist[i].UpdateChapter = update
					novellist[i].UpdateChapter.IsNew = true
					fmt.Printf("Update Novel Chapter: [ %s ] %s %s %s %s\n", item.Title, update.Title, update.Link, update.Time.Format("2006-01-02 15:04:05"), update.MD5)
				}
			}
			sort.Sort(novellist)
			fmt.Println(novellist)
			UpdateTime = time.Now()
			fmt.Printf("Update at:%s\n", UpdateTime.Format("2006-01-02 15:04:05"))
			break
		}
	}
}

func ToRSS(novellist NovelList) []byte {
	createdTime, err := time.Parse(time.RFC3339, "2017-07-02T15:14:00+08:00") // fixed
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("RSS Updated at:%s\n", UpdateTime.Format("2006-01-02 15:04:05"))
	feed := &feeds.Feed{
		Title:       "Chinese Novel RSS",
		Link:        &feeds.Link{Href: "http://140.138.152.121:14112"},
		Description: "Generate Chinese Novels to RSS Feed",
		Author:      &feeds.Author{Name: "fake<fake@fake.org>", Email: "fake@fake.org"},
		Created:     createdTime,
	}
	for i, item := range novellist {
		if novellist[i].UpdateChapter.IsNew == true {
			feed.Add(&feeds.Item{
				Title:       "[" + item.Title + "] " + novellist[i].UpdateChapter.Title,
				Link:        &feeds.Link{Href: novellist[i].UpdateChapter.Link},
				Description: item.Title,
				Id:          novellist[i].UpdateChapter.Link,
				Author:      &feeds.Author{Name: "fake<fake@fake.org>", Email: "fake@fake.org"},
				Created:     novellist[i].UpdateChapter.Time,
			})
			//novellist[i].UpdateChapter.IsNew = false
		}
	}
	rss, err := feed.ToRss()
	if err != nil {
		log.Fatal(err)
		return nil
	} else {
		return []byte(rss)
	}
}
