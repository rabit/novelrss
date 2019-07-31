package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/djimenez/iconv-go"
)

type NovelItem struct {
	Title       string `json:"title"`
	Url         string `json:"url"`
	UpdateTitle string
	UpdateUrl   string
}

type NovelSite struct {
	Selector string
	BaseUrl  string
}

var selectorList = map[string]NovelSite{
	"uukanshu": {Selector: ".zuixin", BaseUrl: "http://www.uukanshu.net"},
	"qidian":   {Selector: ".update", BaseUrl: "https:"},
	"shuqu8":   {Selector: ".stats .fl", BaseUrl: "http://www.shuqu8.com/"},
}

func getUpdateChapter(novelUrl string) (utf8Title string, targetUrl string) {

	doc, err := goquery.NewDocument(novelUrl)
	if err != nil {
		log.Fatal(err)
	}

	selector := ".zuixin"
	targetUrl = "http://"
	for k, v := range selectorList {
		if strings.Contains(novelUrl, k) {
			selector = v.Selector
			targetUrl = v.BaseUrl
		}
	}
	utf8Title = "not found"
	// Find the review items
	fmt.Printf("=== selecor (%s)===\n", selector)
	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		title := s.Find("a").Text()
		link, _ := s.Find("a").Attr("href")
		fmt.Printf("title:%s link: %s\n", title, link)
		if utf8.ValidString(title) == false {
			utf8Title, err = iconv.ConvertString(title, "gb2312", "utf-8")
			if err != nil {
				fmt.Println(err.Error())
				utf8Title = "not found"
			}
		} else {
			utf8Title = title
		}
		if link[0] != '/' {
			targetUrl = novelUrl + link
		} else {
			targetUrl = targetUrl + link
		}
		fmt.Printf("%d: %s %s\n", i, utf8Title, targetUrl)
	})
	return utf8Title, targetUrl
}

func ScrapUpdate(novellist *[]NovelItem) {

	for i, item := range *novellist {
		newtitle, newurl := getUpdateChapter(item.Url)
		(*novellist)[i].UpdateTitle = newtitle
		(*novellist)[i].UpdateUrl = newurl
	}
	fmt.Printf("Update at:%s\n", time.Now().Format("2006-01-02 15:04:05"))
}

func main() {
	noveljsfile := os.Args[1]
	jsondata, err := ioutil.ReadFile(noveljsfile)
	if err != nil {
		fmt.Println(err.Error())
	}

	novelList := make([]NovelItem, 0)
	err = json.Unmarshal(jsondata, &novelList)
	if err != nil {
		fmt.Println("parse json error", err.Error())
		return
	}
	fmt.Println("load novel data: ", noveljsfile)
	for k, v := range novelList {
		fmt.Printf("%d %s %s\n", k, v.Title, v.Url)
	}

	ScrapUpdate(&novelList)
}
