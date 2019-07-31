package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/rabit/novelrss"
)

func main() {
	port := flag.String("p", "14112", "port to serve on")
	noveljsfile := flag.String("f", "novel.js", "the novel json data")
	flag.Parse()

	fmt.Println("web port: ", *port)
	jsondata, err := ioutil.ReadFile(*noveljsfile)
	if err != nil {
		log.Fatal(err.Error())
	}

	var novellist novelrss.NovelList
	err = json.Unmarshal(jsondata, &novellist)
	if err != nil {
		fmt.Println("parse json error", err.Error())
		return
	}
	fmt.Println("load novel data: ", *noveljsfile)
	for k, v := range novellist {
		fmt.Printf("%d %s %s\n", k, v.Title, v.Url)
	}

	go novelrss.Update(novellist)

	// gin
	r := gin.Default()
	r.Use(static.Serve("/statics", novelrss.WebAssetsDir("./web/statics")))
	t, err := novelrss.LoadTemplate("./web/templates")
	if err != nil {
		panic(err)
	}
	r.SetHTMLTemplate(t)
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{"title": "追書", "novels": novellist})
	})
	r.GET("/rss", func(c *gin.Context) {
		//c.Data(http.StatusOK, "application/rss+xml", novelrss.ToRSS(novellist))
		c.Data(http.StatusOK, "text/xml;charset=UTF-8", novelrss.ToRSS(novellist))
	})
	r.Run(":" + *port)
}
