package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

var (
	domain string
	port   string
)

func runHTTPServer() {
	flag.StringVar(&domain, "domain", "www.example.com", "domain whitelist for configuring autocert")
	flag.StringVar(&port, "port", ":443", "custom port for server to listen on")
	flag.Parse()

	http.HandleFunc("/calendar/mrha/", handleMrhaCalendar)

	manager := &autocert.Manager{
		Cache:      autocert.DirCache("secrets"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/calendar/mrha/", handleMrhaCalendar)

	server := &http.Server{
		Addr:         port,
		TLSConfig:    manager.TLSConfig(),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Println("listening on", port)
	log.Fatal(server.ListenAndServeTLS("", ""))
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
