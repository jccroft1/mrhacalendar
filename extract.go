package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

func fetch(url string) (*html.Node, error) {
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
	return doc, err
}

func findFixturesTable(d *html.Node) (*html.Node, error) {
	var table *html.Node
	f := func(n *html.Node) bool {
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
	forEachNode(d, f)
	if table == nil {
		return nil, fmt.Errorf("could not find table")
	}
	return table, nil
}

func parseTable(t *html.Node) ([][]string, error) {
	var data [][]string
	var row []string
	parse := func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "tr" {
			if len(row) > 0 {
				data = append(data, row)
			}
			row = []string{}
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
			row = append(row, text)
		}
		return false
	}
	forEachNode(t, parse)

	if len(data) == 0 {
		return nil, fmt.Errorf("could not find any fixtures")
	}
	return data, nil

}

func createCalendar(teamID string, fixtures []FixtureData) (*Calendar, error) {
	ical := Calendar{
		Name: "MRHA Fixtures " + teamID,
	}
	loc, _ := time.LoadLocation("Europe/London")
	layouts := []string{"2 Jan 2006 15:04pm", "2 Jan 2006 15.04pm", "2 Jan 2006 15.04", "2 Jan 2006 15:04"}
	for i, f := range fixtures {
		var start time.Time
		var err error
		for _, l := range layouts {
			start, err = time.ParseInLocation(l, f.Date+" "+f.Time, loc)
			if err == nil {
				break
			}
		}
		if err != nil {
			start, err = time.ParseInLocation("2 Jan 2006", f.Date, loc)
		}
		end := start.Add(time.Hour * 2)
		ical.Add(Event{
			Name:        f.Home + " vs " + f.Away,
			Created:     time.Now(),
			UID:         "fixture" + strconv.Itoa(i),
			Start:       start,
			End:         end,
			Location:    f.Venue,
			Description: f.Score,
		})
	}

	return &ical, nil
}

// Extract makes an HTTP GET request to the specified URL, parses
// the response as HTML, and returns the links in the HTML document.
func extract(teamID string) (*Calendar, error) {
	doc, err := fetch(fmt.Sprintf(mrhaTeamURL, teamID))
	if err != nil {
		return nil, err
	}

	table, err := findFixturesTable(doc)
	if err != nil {
		return nil, err
	}

	tableData, err := parseTable(table)
	if err != nil {
		return nil, err
	}

	fixtures := make([]FixtureData, 0, len(tableData))
	for _, f := range tableData {
		if len(f) == 6 {
			fixtures = append(fixtures, FixtureData{
				Date:  f[0],
				Home:  f[1],
				Score: f[2],
				Away:  f[3],
				Time:  f[4],
				Venue: f[5],
			})
		}
	}

	return createCalendar(teamID, fixtures)
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
