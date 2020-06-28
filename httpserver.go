package main

import (
	"fmt"
	"log"
	"net/http"
)

func runHTTPServer() {
	http.HandleFunc("/calendar/mrha/", handleMrhaCalendar)
	log.Println("listening on 59463")
	panic(http.ListenAndServe(":59463", nil))
}

func handleMrhaCalendar(w http.ResponseWriter, r *http.Request) {
	log.Println("request", r.URL)
	teamID := r.URL.Query().Get("teamId")
	c, err := cache.Get(teamID)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Fprint(w, c)
}
