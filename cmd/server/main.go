package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"jievow-calendar/api"
	"jievow-calendar/calendar"
)

func main() {
	port := flag.String("port", "", "listen port (default: PORT env or 8080)")
	flag.Parse()

	addr := ":" + resolvePort(*port)
	fmt.Printf("Calendar API listening on %s\n", addr)

	store, err := calendar.LoadStore("data")
	if err != nil {
		log.Fatalf("failed to load data: %v", err)
	}
	log.Printf("Loaded %d records (version %s)", store.Len(), store.Version())

	mux := http.NewServeMux()
	mux.Handle("GET /api/v1/date/{date}", api.NewHandler(store))

	handler := api.CORS(mux)
	log.Fatal(http.ListenAndServe(addr, handler))
}

func resolvePort(flagPort string) string {
	if flagPort != "" {
		return flagPort
	}
	if envPort := os.Getenv("PORT"); envPort != "" {
		return envPort
	}
	return "8080"
}