package main

import (
	"flag"
	"log"
	"net/http"

	"task18/internal/calendar"
	"task18/internal/httpapi"
	"task18/internal/middleware"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	svc := calendar.NewService()
	api := &httpapi.Server{Svc: svc}

	mux := http.NewServeMux()
	mux.HandleFunc("/create_event", api.CreateEvent)
	mux.HandleFunc("/update_event", api.UpdateEvent)
	mux.HandleFunc("/delete_event", api.DeleteEvent)
	mux.HandleFunc("/events_for_day", api.EventsForDay)
	mux.HandleFunc("/events_for_week", api.EventsForWeek)
	mux.HandleFunc("/events_for_month", api.EventsForMonth)

	h := middleware.Logging(mux)
	log.Printf("listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, h))
}
