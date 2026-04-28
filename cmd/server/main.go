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

	flowerStore, err := calendar.LoadFlowerStore("data")
	if err != nil {
		log.Fatalf("failed to load flower data: %v", err)
	}
	log.Printf("Loaded flower data for %d provinces", len(flowerStore.ListProvinces()))

	mux := http.NewServeMux()
	h := api.NewHandler(store, flowerStore)
	mux.Handle("GET /api/v1/date/{date}", h)
	mux.HandleFunc("GET /api/v1/range", h.HandleRange)
	mux.HandleFunc("GET /api/v1/solar-terms", h.HandleSolarTerms)
	mux.HandleFunc("GET /api/v1/flowers", h.HandleFlowers)

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