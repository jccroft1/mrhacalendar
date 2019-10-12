package main

import (
	"fmt"
	"strings"
	"time"
)

// Calendar struct to store info
type Calendar struct {
	Name   string
	Events []Event
}

func (c Calendar) String() string {
	text := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//hacksw/handcal//NONSGML v1.0//EN
NAME:%v
%v
END:VCALENDAR`
	var eventText string
	for _, e := range c.Events {
		eventText = eventText + fmt.Sprint(e)
	}

	return fmt.Sprintf(text, c.Name, strings.TrimSpace(eventText))
}

// Add an Event to the Calendar
func (c *Calendar) Add(e Event) {
	c.Events = append(c.Events, e)
}

// Event struct for icalendar data
type Event struct {
	Name        string
	Created     time.Time
	UID         string
	Start       time.Time
	End         time.Time
	Location    string
	Description string
}

func (e Event) String() string {
	layout := "20060102T150405"
	text := `BEGIN:VEVENT
UID:%v
DTSTAMP;TZID=Europe/London:%v
DTSTART;TZID=Europe/London:%v
DTEND;TZID=Europe/London:%v
SUMMARY:%v
LOCATION:%v
DESCRIPTION:%v
END:VEVENT
`

	return fmt.Sprintf(text, e.UID, e.Created.Format(layout), e.Start.Format(layout), e.End.Format(layout), e.Name, e.Location, e.Description)
}
