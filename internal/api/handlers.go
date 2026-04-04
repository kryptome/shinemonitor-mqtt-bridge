package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kryptome/shinemonitor-mqtt-bridge/internal/cache"
	"github.com/kryptome/shinemonitor-mqtt-bridge/internal/shinemonitor"
)

type Server struct {
	client *shinemonitor.Client
}

func NewServer(client *shinemonitor.Client) *Server {
	return &Server{client: client}
}

func (s *Server) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/status", s.handleCached("status", func() (interface{}, error) { return s.client.GetStatus() }))
	r.Get("/now", s.handleCached("now", func() (interface{}, error) { return s.client.GetEnergyNow() }))
	r.Get("/summary", s.handleCached("summary", func() (interface{}, error) { return s.client.GetEnergySummary() }))
	r.Get("/dashboard", s.handleCached("dashboard", func() (interface{}, error) { return s.client.GetWebQueryPlants() }))
	
	r.Get("/timeline", s.handleTimeline("day"))
	r.Get("/timeline/month", s.handleTimeline("month"))
	r.Get("/timeline/year", s.handleTimeline("year"))
	r.Get("/timeline/total", s.handleTimeline("total"))
	
	r.Get("/plant", s.handlePlant)

	return r
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
	}
}

func (s *Server) handleCached(key string, fetcher func() (interface{}, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if val, ok := cache.Get(key); ok {
			writeJSON(w, val)
			return
		}

		res, err := fetcher()
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			writeJSON(w, map[string]string{"Error": err.Error()})
			return
		}
		writeJSON(w, res)
	}
}

func isToday(dateStr string) bool {
	if dateStr == "" {
		return true // Missing date defaults to today
	}
	todayStr := time.Now().Format("2006-01-02")
	return dateStr == todayStr
}

func (s *Server) handleTimeline(mode string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		date := r.URL.Query().Get("date")
		if date == "" && mode != "total" {
			if mode == "day" {
				date = time.Now().Format("2006-01-02")
			} else if mode == "month" {
				date = time.Now().Format("2006-01")
			} else if mode == "year" {
				date = time.Now().Format("2006")
			}
		}

		cacheKey := "timeline_" + mode + "_" + date
		var ttl = 300 * time.Second

		useCache := true
		if mode == "day" && isToday(date) {
			useCache = false
		} else if mode == "month" && date == time.Now().Format("2006-01") {
			useCache = false
		} else if mode == "year" && date == time.Now().Format("2006") {
			useCache = false
		}

		if useCache {
			if val, ok := cache.Get(cacheKey); ok {
				writeJSON(w, val)
				return
			}
		}

		var res []shinemonitor.EnergyTimelineResponse
		var err error

		switch mode {
		case "day":
			// Mapped strictly to new DeviceChart API per user fix
			res, err = s.client.GetDeviceChart(date)
		case "month":
			// Placeholder behavior preserving previous signatures if still supported
			res, err = s.client.GetEnergyTimelineMonth(date)
		case "year":
			res, err = s.client.GetEnergyTimelineYear(date)
		case "total":
			res, err = s.client.GetEnergyTimelineTotal()
			ttl = 3600 * time.Second
		}

		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			writeJSON(w, map[string]string{"Error": err.Error()})
			return
		}

		if useCache {
			cache.Set(cacheKey, res, ttl)
		}
		writeJSON(w, res)
	}
}

func (s *Server) handlePlant(w http.ResponseWriter, r *http.Request) {
	cacheKey := "plant"
	if val, ok := cache.Get(cacheKey); ok {
		writeJSON(w, val)
		return
	}

	res, err := s.client.GetPlantInfo()
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		writeJSON(w, map[string]string{"Error": err.Error()})
		return
	}

	cache.Set(cacheKey, res, 3600*time.Second)
	writeJSON(w, res)
}
