package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	ResultSinceYear  = 2006
	ResultSinceMonth = 4
)

const (
	JapanShogiAssociationHostName = "www.shogi.or.jp"
)

var targetMonth string

func init() {
	flag.StringVar(&targetMonth, "m", "", "Scrape monthly results (e.g. \"201804\")")
}

type Match struct {
	MatchName    string `json:"matchName"`
	BeginDate    string `json:"beginDate"`
	EndDate      string `json:"endDate"`
	FirstPlayer  Player `json:"firstPlayer"`
	SecondPlayer Player `json:"secondPlayer"`
	Note         string `json:"note,omitempty"`
}

type Player struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Result string `json:"result"`
}

func scrape(html io.Reader) []Match {
	doc, err := goquery.NewDocumentFromReader(html)
	if err != nil {
		log.Fatal(err)
	}

	year := parseYear(doc)
	matches := []Match{}
	var beginDate, endDate string
	matchName := ""

	// Find the review items
	table := doc.Find(".tableElements01 tbody").First()
	table.Children().Each(func(i int, s *goquery.Selection) {
		if isDateRow(s) {
			beginDate, endDate = parseDate(s, year)
			return
		}
		if includeMatchName(s) {
			matchName = parseMatchName(s)
		}

		m := parseMatch(s, matchName, beginDate, endDate)
		matches = append(matches, m)
	})

	return matches
}

func parseYear(doc *goquery.Document) (year int) {
	str := doc.Find(".headingElementsA01").Text()
	_, err := fmt.Sscanf(str, "%d年", &year)
	if err != nil {
		log.Fatalln(err)
	}
	return
}

func isDateRow(s *goquery.Selection) bool {
	return s.Children().Length() == 1
}

func parseDate(s *goquery.Selection, year int) (string, string) {
	str := s.Text()

	var month, beginDay, endDay int
	_, err := fmt.Sscanf(str, "%d月%d日", &month, &beginDay)
	if err != nil {
		_, err = fmt.Sscanf(str, "%d月%d・%d日", &month, &beginDay, &endDay)
	} else {
		endDay = beginDay
	}

	beginDate := fmt.Sprintf("%d/%02d/%02d", year, month, beginDay)
	endDate := fmt.Sprintf("%d/%02d/%02d", year, month, endDay)
	return beginDate, endDate
}

func includeMatchName(s *goquery.Selection) bool {
	return s.Children().Length() == 6
}

func parseMatchName(s *goquery.Selection) string {
	str := s.Children().First().Text()
	return strings.TrimSpace(str)
}

func parseMatch(s *goquery.Selection, matchName, beginDate, endDate string) Match {
	row := s.Find("td.tac")
	if row.Length() != 4 {
		log.Fatalf("Unexpected html structure at parseMatch, %d", row.Length())
	}

	firstPlayerResult := parseResult(row.Eq(0))
	firstPlayer := parsePlayer(row.Eq(1), firstPlayerResult)

	secondPlayerResult := parseResult(row.Eq(3))
	secondPlayer := parsePlayer(row.Eq(2), secondPlayerResult)

	note := parseNote(s)

	return Match{matchName, beginDate, endDate, firstPlayer, secondPlayer, note}
}

func parseResult(s *goquery.Selection) string {
	switch s.Text() {
	case "○":
		return "win"
	case "●":
		return "lose"
	case "□":
		return "win without playing"
	case "■":
		return "lose without playing"
	default:
		return "unknown"
	}
}

func parsePlayer(s *goquery.Selection, result string) Player {
	name := s.Text()

	var id string
	href, exists := s.Find("a").Attr("href")
	if exists {
		href = strings.TrimPrefix(href, "/player/")
		id = strings.TrimSuffix(href, ".html")
	} else {
		id = "n/a"
	}
	return Player{id, name, result}
}

func parseNote(s *goquery.Selection) string {
	n := s.Children().Last()
	note := n.Text()
	href, exists := n.Find("a").Attr("href")
	if exists {
		note = note + " " + sanitizeURL(href)
	}
	return note
}

func sanitizeURL(str string) string {
	u, err := url.Parse(str)
	if err != nil {
		log.Fatal(err)
	}

	if u.IsAbs() {
		return str
	} else {
		u.Scheme = "https"
		u.Host = JapanShogiAssociationHostName
		return u.String()
	}
}

func ScrapeFromDate(year, month int) {
	url := fmt.Sprintf("https://www.shogi.or.jp/game/result/%d%02d.html", year, month)
	out := fmt.Sprintf("results/%d%02d.json", year, month)

	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	matches := scrape(res.Body)
	jsonBytes, err := json.MarshalIndent(matches, "", "  ")
	if err != nil {
		log.Fatalln("JSON Marshal error:", err)
	}

	f, err := os.Create(out)
	if err != nil {
		log.Fatal(err)
	}
	f.Write(jsonBytes)
	fmt.Println("Output: ", out)
}

func ScrapeAllResult() {
	now := time.Now()
	t := time.Date(ResultSinceYear, ResultSinceMonth, 1, 0, 0, 0, 0, time.UTC)

	for t.Before(now) {
		ScrapeFromDate(t.Year(), int(t.Month()))
		t = t.AddDate(0, 1, 0)
	}
}

func ScrapeMonthlyResult() {
	const layout = "200601"
	t, err := time.Parse(layout, targetMonth)
	if err != nil {
		log.Fatal(err)
	}

	ScrapeFromDate(t.Year(), int(t.Month()))
}

func main() {
	flag.Parse()

	if targetMonth == "" {
		ScrapeAllResult()
	} else {
		ScrapeMonthlyResult()
	}
}
