package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kryptome/shinemonitor-mqtt-bridge/internal/cache"
	"github.com/kryptome/shinemonitor-mqtt-bridge/internal/mqtt"
	"github.com/kryptome/shinemonitor-mqtt-bridge/internal/shinemonitor"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	client *shinemonitor.Client
	mq     *mqtt.Client
}

func NewServer(client *shinemonitor.Client, mq *mqtt.Client) *Server {
	return &Server{
		client: client,
		mq:     mq,
	}
}

func (s *Server) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/status", s.handleStatus)
	r.Get("/now", s.handleNow)
	r.Get("/summary", s.handleSummary)
	r.Get("/dashboard", s.handleDashboard)
	
	r.Get("/timeline", s.handleTimeline("day"))
	r.Get("/timeline/month", s.handleTimeline("month"))
	r.Get("/timeline/year", s.handleTimeline("year"))
	r.Get("/timeline/total", s.handleTimeline("total"))
	
	r.Get("/plant", s.handlePlant)

	r.Post("/forceDiscovery", s.handleForceDiscovery)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	return r
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
	}
}

// @Summary Get device status
// @Description Fetches the current online/offline status of the inverter
// @Tags Live Data
// @Produce json
// @Success 200 {object} shinemonitor.StatusResponse
// @Router /status [get]
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if val, ok := cache.Get("status"); ok {
		writeJSON(w, val)
		return
	}

	res, err := s.client.GetStatus()
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		writeJSON(w, map[string]string{"Error": err.Error()})
		return
	}
	writeJSON(w, res)
}

// @Summary Get current power
// @Description Gets the immediate live power output of the plant
// @Tags Live Data
// @Produce json
// @Success 200 {object} shinemonitor.EnergyNowResponse
// @Router /now [get]
func (s *Server) handleNow(w http.ResponseWriter, r *http.Request) {
	if val, ok := cache.Get("now"); ok {
		writeJSON(w, val)
		return
	}
	res, err := s.client.GetEnergyNow()
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		writeJSON(w, map[string]string{"Error": err.Error()})
		return
	}
	writeJSON(w, res)
}

// @Summary Get energy summary
// @Description Gets the today, month, year, and total energy production
// @Tags Live Data
// @Produce json
// @Success 200 {object} shinemonitor.EnergySummaryResponse
// @Router /summary [get]
func (s *Server) handleSummary(w http.ResponseWriter, r *http.Request) {
	if val, ok := cache.Get("summary"); ok {
		writeJSON(w, val)
		return
	}
	res, err := s.client.GetEnergySummary()
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		writeJSON(w, map[string]string{"Error": err.Error()})
		return
	}
	writeJSON(w, res)
}

// @Summary Get detailed dashboard
// @Description Provides full metadata and raw readings about the plant including efficiency and cfValue
// @Tags Live Data
// @Produce json
// @Success 200 {object} shinemonitor.WebQueryPlant
// @Router /dashboard [get]
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if val, ok := cache.Get("dashboard"); ok {
		writeJSON(w, val)
		return
	}
	res, err := s.client.GetWebQueryPlants()
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		writeJSON(w, map[string]string{"Error": err.Error()})
		return
	}
	writeJSON(w, res)
}

func isToday(dateStr string) bool {
	if dateStr == "" {
		return true // Missing date defaults to today
	}
	return dateStr == time.Now().Format("2006-01-02")
}

// @Summary Get energy timeline
// @Description Fetches historical energy production charts for day, month, year, or total metrics
// @Tags Historical
// @Produce json
// @Param date query string false "Date for historical records depending on timeframe (YYYY-MM-DD, YYYY-MM, or YYYY)"
// @Success 200 {array} shinemonitor.EnergyTimelineResponse
// @Router /timeline [get]
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
			res, err = s.client.GetDeviceChart(date)
		case "month":
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

// @Summary Get plant details
// @Description Grabs static configuration, address and nominal metrics for the plant
// @Tags Meta
// @Produce json
// @Success 200 {object} shinemonitor.PlantInfoResponse
// @Router /plant [get]
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

// @Summary Force resend Home Assistant discovery
// @Description Manually triggers the re-publishing of MQTT discovery payloads for Home Assistant
// @Tags System
// @Success 200 {object} map[string]string
// @Router /forceDiscovery [post]
func (s *Server) handleForceDiscovery(w http.ResponseWriter, r *http.Request) {
	if s.mq == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		writeJSON(w, map[string]string{"Error": "MQTT client not initialized"})
		return
	}

	s.mq.PublishDiscovery()
	writeJSON(w, map[string]string{"status": "Discovery payloads sent"})
}
