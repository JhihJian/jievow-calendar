package main

import (
	"fmt"
	"log"
	"net/http"

	"jievow-calendar/api"
	"jievow-calendar/calendar"
)

func main() {
	store, err := calendar.LoadStore("data")
	if err != nil {
		log.Fatalf("failed to load data: %v", err)
	}
	log.Printf("Loaded %d records (version %s)", store.Len(), store.Version())

	mux := http.NewServeMux()
	mux.Handle("GET /api/v1/date/{date}", api.NewHandler(store))

	handler := api.CORS(mux)

	addr := ":8080"
	fmt.Printf("Calendar API listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}