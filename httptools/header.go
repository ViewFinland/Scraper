package httptools

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

type Generator struct {
	browserAgents userAgents
}

type userAgents []struct {
	Useragent string `json:"useragent"`
}

func NewGenerator() *Generator {
	generator := &Generator{}
	data, err := ioutil.ReadFile("useragents.json")
	if err != nil {
		panic(err)
	}

	var agents userAgents
	err = json.Unmarshal(data, &agents)
	if err != nil {
		panic(err)
	}

	generator.browserAgents = agents
	return generator
}

func (g *Generator) AddHeaders(header *http.Header) {
	header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")
	header.Add("Accept-Charset", "utf-8")
	header.Add("Accept-Language", "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7")
	header.Add("Cache-Control", "no-cache")
	header.Add("Content-Type", "application/json; charset=utf-8")
	header.Add("User-Agent", g.GetRandomUserAgent())
}

func (g *Generator) GetRandomUserAgent() string {
	rand.Seed(time.Now().Unix())
	return g.browserAgents[rand.Intn(len(g.browserAgents))].Useragent
}
