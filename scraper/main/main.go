package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
	_ "github.com/lib/pq"
	"github.com/m-lukas/viewfinland/httptools"
)

type Post struct {
	ID         string
	ImageURL   string
	HashTags   string
	Location   string
	Likes      int
	UserHandle string
}

type Scraper struct {
	db      *sql.DB
	tags    []string
	counter int
}

func main() {
	scraper := &Scraper{
		tags: []string{
			"sumuinen",
			"nuuksioclassic",
			"nuuksiobikepark",
			"nuuksiosma",
			"nuuksio70",
		},
	}

	postgresHost := "127.0.0.1"
	connectionString := fmt.Sprintf("host=%s user=postgres dbname=postgres sslmode=disable", postgresHost)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		panic(err)
	}

	scraper.db = db

	for _, tag := range scraper.tags {
		scraper.queryTag(tag)
	}
}

func (s *Scraper) queryTag(tag string) error {
	fmt.Printf("Querying tag: %s\n", tag)

	url := fmt.Sprintf("http://picpanzee.com/tag/%s", tag)

	generator := httptools.NewGenerator()

	c := colly.NewCollector()

	c.UserAgent = generator.GetRandomUserAgent()
	c.OnRequest(func(r *colly.Request) {
		generator.AddHeaders(r.Headers)
	})
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("err occured in tag scraper: %v\n", err)
	})
	c.OnResponse(func(r *colly.Response) {
		log.Println("Processed post!")
	})

	c.OnHTML("a.grid-media-media", func(e *colly.HTMLElement) {
		s.counter++

		postDetailURL := e.Attr("href")
		fmt.Printf("Visiting %s\n", postDetailURL)
		post, err := s.getPostDetail(postDetailURL)
		if err != nil {
			log.Printf("ignored err occured in tag scraper: %v\n", err)
		}

		err = s.insertPost(post)
		if err != nil {
			log.Printf("ignored err occured in tag scraper: %v\n", err)
		}
	})

	err := c.Visit(url)
	if err != nil {
		return err
	}

	fmt.Println("Finished!")
	return nil
}

func (s *Scraper) getPostDetail(url string) (*Post, error) {
	post := &Post{}
	d := colly.NewCollector()

	generator := httptools.NewGenerator()

	d.UserAgent = generator.GetRandomUserAgent()
	d.OnRequest(func(r *colly.Request) {
		generator.AddHeaders(r.Headers)
	})
	d.OnResponse(func(r *colly.Response) {
		log.Println(fmt.Sprintf("Processed post %v!", s.counter))
	})
	d.OnError(func(r *colly.Response, err error) {
		log.Printf("err occured in detail scraper: %v\n", err)
	})

	d.OnHTML(".media-single-media", func(e *colly.HTMLElement) {
		imageURL := e.Attr("src")
		hashTags := e.Attr("alt")

		post.ImageURL = imageURL
		post.HashTags = hashTags
	})

	d.OnHTML(".media-location a", func(e *colly.HTMLElement) {
		location := e.Attr("title")

		post.Location = location
	})

	d.OnHTML(".media-single-title", func(e *colly.HTMLElement) {
		likes := e.ChildText("i")
		strings.ReplaceAll(likes, " ", "")
		strings.ReplaceAll(likes, "Likes", "")
		strings.ReplaceAll(likes, "Like", "")

		likesCount, err := strconv.Atoi(likes)
		if err != nil {
			likesCount = 0
		}

		post.Likes = likesCount
	})

	d.OnHTML(".media-location a", func(e *colly.HTMLElement) {
		location := e.Attr("title")

		post.Location = location
	})

	d.OnHTML("h4.media-heading a", func(e *colly.HTMLElement) {
		userHandle := e.Attr("title")

		post.UserHandle = userHandle
	})

	urlParts := strings.Split(url, "/")
	id := urlParts[len(urlParts)-1]
	post.ID = id

	err := d.Visit(url)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (s *Scraper) insertPost(p *Post) error {
	_, err := s.db.Exec("INSERT INTO posts(ID, ImageURL, HashTags, Location, Likes, UserHandle) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(ID) DO NOTHING",
		p.ID,
		p.ImageURL,
		p.HashTags,
		p.Location,
		p.Likes,
		p.UserHandle,
	)

	if err != nil {
		return err
	}

	return nil
}
