package main

import (
	"fmt"
	"log"
	"net/http"
)

func runHTTPServer() {
	http.HandleFunc("/calendar/mrha/", handleMrhaCalendar)
	log.Println("Listening on 59463")
	panic(http.ListenAndServe(":59463", nil))
}

func handleMrhaCalendar(w http.ResponseWriter, r *http.Request) {
	log.Println("handling request ", r.URL)
	teamID := r.URL.Query().Get("teamId")
	c := get(teamID)

	var err error
	if c == nil {
		fmt.Println("extracting", teamID)
		c, err = extract(teamID)

		if err != nil {
			return
		}
		// no need to let storing in cache block current request
		go set(teamID, c)
	} else {
		fmt.Println("Using cache")
	}

	fmt.Fprint(w, c)
}
