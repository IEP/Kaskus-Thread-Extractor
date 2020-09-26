package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/gocolly/colly/v2"
)

// Story store thread content
type Story struct {
	Username  string
	PostCount int
	Headline  string
	Post      string
}

func main() {
	// Argument parsing
	var baseURL, title string
	flag.StringVar(&baseURL, "url", "", "URL to the thread")
	flag.StringVar(&title, "title", "", "The thread title")
	flag.Parse()

	if baseURL == "" || title == "" {
		panic("No parameter passed")
	}

	thread := colly.NewCollector(colly.Async())
	story := make(chan Story, 501*20)
	var ts string

	// First post
	thread.OnHTML("div.postItemFirst", func(e *colly.HTMLElement) {
		username := e.ChildText("a[href*='profile'][itemprop='url']")
		post := e.ChildText("article")
		ts = username
		story <- Story{
			username, 0, "", post,
		}
	})

	// Comments
	thread.OnHTML("div[itemprop='comment']", func(e *colly.HTMLElement) {
		username := e.ChildText("a[href*='profile'][itemprop='url']")
		post := e.ChildText("article")
		postCountTxt := e.ChildAttr("a[id*='postcount']", "name")
		postCount, _ := strconv.ParseInt(postCountTxt, 10, 64)
		headline := e.ChildText("h1[itemprop='headline']")

		story <- Story{
			username, int(postCount), headline, post,
		}
	})

	// Pagination
	thread.OnHTML("a[href*='order=asc'][href*='thread']", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		thread.Visit(e.Request.AbsoluteURL(link))
	})

	thread.Visit(baseURL)

	go func() {
		thread.Wait()
		close(story)
	}()

	// Wait for TS to be assigned
	for {
		if ts != "" {
			break
		}
	}

	for storyContent := range story {
		if storyContent.Username == ts {
			content := []byte(storyContent.Post)
			headline := storyContent.Headline
			if headline == "" {
				headline = "NONE"
			}
			err := ioutil.WriteFile(fmt.Sprintf("result/%s-%d-%s.txt", title, storyContent.PostCount, headline), content, 0644)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
