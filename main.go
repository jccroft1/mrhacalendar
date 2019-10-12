package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

// FixtureData stores information for a single fixture
type FixtureData struct {
	Date  string
	Time  string
	Home  string
	Away  string
	Score string
	Venue string
}

var mrhaTeamURL = "http://www.mrha.co.uk/mwTeam.asp?teamid=%v"

var cache = struct {
	sync.RWMutex
	data map[string]*Calendar
}{
	data: make(map[string]*Calendar),
}

func main() {

	http.HandleFunc("/calendar/mrha/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("handling request ", r.URL)
		teamID := r.URL.Query().Get("teamId")
		cache.RLock()
		c := cache.data[teamID]
		cache.RUnlock()

		var err error
		if c == nil {
			url := fmt.Sprintf(mrhaTeamURL, teamID)
			c, err = Extract(url)
			c.Name = "MRHA Fixtures " + teamID
			if err != nil {
				return
			}
			cache.Lock()
			cache.data[teamID] = c
			cache.Unlock()
		} else {
			fmt.Println("Using cache")
		}

		fmt.Fprint(w, c)
	})
	log.Println("Listening on 59463")
	panic(http.ListenAndServe(":59463", nil))
}

// Extract makes an HTTP GET request to the specified URL, parses
// the response as HTML, and returns the links in the HTML document.
func Extract(url string) (*Calendar, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("getting %s: %s", url, resp.Status)
	}

	doc, err := html.Parse(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("parsing %s as HTML: %v", url, err)
	}

	fmt.Println("html parsed")

	var table *html.Node
	findFixturesTable := func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "table" {
			for _, c := range n.Attr {
				if c.Key != "name" || c.Val != "fixturestable" {
					continue
				}
				table = n
				return true
			}
		}
		return false
	}
	forEachNode(doc, findFixturesTable)
	if table == nil {
		return nil, fmt.Errorf("could not find table")
	}

	var fixturesRaw [][]string
	var fixture []string
	extractFixtures := func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "tr" {
			if len(fixture) > 0 {
				fixturesRaw = append(fixturesRaw, fixture)
			}
			fixture = []string{}
		}
		if n.Type == html.ElementNode && n.Data == "td" {
			var text string
			extractText := func(node *html.Node) bool {
				if node.FirstChild == nil {
					text += strings.TrimSpace(node.Data)
				}
				return false
			}
			forEachNode(n, extractText)
			fixture = append(fixture, text)
		}
		return false
	}
	forEachNode(table, extractFixtures)

	fixtures := make([]FixtureData, 0, len(fixturesRaw))
	for _, f := range fixturesRaw {
		if len(f) == 6 {
			fixtures = append(fixtures, FixtureData{
				Date:  f[0],
				Time:  f[4],
				Home:  f[1],
				Away:  f[3],
				Score: f[2],
				Venue: f[5],
			})
		}
	}

	ical := Calendar{}
	loc, _ := time.LoadLocation("Europe/London")
	for i, f := range fixtures {
		start, err := time.ParseInLocation("2 Jan 2006 15:04pm", f.Date+" "+f.Time, loc)
		if err != nil {
			start, err = time.ParseInLocation("2 Jan 2006 15.04pm", f.Date+" "+f.Time, loc)
		}
		if err != nil {
			start, err = time.ParseInLocation("2 Jan 2006 15.04", f.Date+" "+f.Time, loc)
		}
		if err != nil {
			start, err = time.ParseInLocation("2 Jan 2006 15:04", f.Date+" "+f.Time, loc)
		}
		if err != nil {
			start, err = time.ParseInLocation("2 Jan 2006", f.Date, loc)
		}
		end := start.Add(time.Hour * 2)
		ical.Add(Event{
			Name:     f.Home + " vs " + f.Away,
			Created:  time.Now(),
			UID:      "fixture" + strconv.Itoa(i),
			Start:    start,
			End:      end,
			Location: f.Venue,
		})
	}

	return &ical, nil
}

//!-Extract

// Copied from gopl.io/ch5/outline2.
func forEachNode(n *html.Node, f func(n *html.Node) bool) {
	if f != nil {
		if f(n) {
			return
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		forEachNode(c, f)
	}
}
