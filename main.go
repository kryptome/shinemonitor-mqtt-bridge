package main

import (
	"log"
	"net/http"
	"time"

	"github.com/kryptome/shinemonitor-mqtt-bridge/internal/api"
	"github.com/kryptome/shinemonitor-mqtt-bridge/internal/cache"
	"github.com/kryptome/shinemonitor-mqtt-bridge/internal/config"
	"github.com/kryptome/shinemonitor-mqtt-bridge/internal/mqtt"
	"github.com/kryptome/shinemonitor-mqtt-bridge/internal/shinemonitor"

	_ "github.com/kryptome/shinemonitor-mqtt-bridge/docs"
)

// @contact.name Piyush Mishra

type InverterState int

const (
	StateUnknown InverterState = iota
	StateOnline
	StateOffline
)

var currentInverterState = StateUnknown

func main() {
	cfg := config.LoadConfig()

	log.Printf("Starting ShineMonitor Bridge (Poll Interval: %v)\n", cfg.PollInterval)

	smClient := shinemonitor.NewClient(cfg)
	
	var mqClient *mqtt.Client
	if cfg.MQTTBroker != "" {
		c, err := mqtt.Connect(cfg)
		if err != nil {
			log.Fatalf("FAILED TO CONNECT TO MQTT: %v", err)
		}
		mqClient = c
		mqClient.PublishDiscovery()

		// Background Discovery Refresh Goroutine (every 6 hours)
		go func() {
			discoveryTicker := time.NewTicker(6 * time.Hour)
			defer discoveryTicker.Stop()
			for range discoveryTicker.C {
				log.Println("Discovery refresh triggered (6h interval)")
				mqClient.PublishDiscovery()
			}
		}()
	}

	// Background Poller Goroutine
	go func() {
		ticker := time.NewTicker(cfg.PollInterval)
		defer ticker.Stop()
		
		doPoll(cfg, smClient, mqClient)

		for range ticker.C {
			doPoll(cfg, smClient, mqClient)
		}
	}()
	
	server := api.NewServer(smClient, mqClient)
	listenAddr := ":" + cfg.Port
	log.Printf("Listening REST API on %s\n", listenAddr)
	if err := http.ListenAndServe(listenAddr, server.Routes()); err != nil {
		log.Fatalf("HTTP Listener Failed: %v\n", err)
	}
}

func doPoll(cfg *config.Config, sm *shinemonitor.Client, mq *mqtt.Client) {
	// Execute a single optimized `webQueryPlants` call
	plant, err := sm.GetWebQueryPlants()
	deviceData, devErr := sm.GetDeviceDataOneDayPaging()
	
	if devErr != nil {
		log.Printf("Warning: Failed to fetch device data: %v", devErr)
	}

	ttl := cfg.PollInterval + 5*time.Second

	if err == nil {
		// Populate all caches individually so existing routes resolve natively
		statusStr := "Offline"
		if plant.Status == 0 {
			statusStr = "Online"
			if currentInverterState != StateOnline {
				log.Println("Inverter came Online, re-publishing discovery...")
				if mq != nil {
					mq.PublishDiscovery()
				}
				currentInverterState = StateOnline
			}
		} else {
			currentInverterState = StateOffline
		}

		cache.Set("status", &shinemonitor.StatusResponse{Status: statusStr}, ttl)

		cache.Set("now", &shinemonitor.EnergyNowResponse{
			TimeStamp:  time.Now(),
			Energy:     plant.OutputPower,
			Unit:       "kW",
			DeviceData: deviceData,
		}, ttl)

		cache.Set("summary", &shinemonitor.EnergySummaryResponse{
			Today: plant.Energy,
			Month: plant.EnergyMonth,
			Year:  plant.EnergyYear,
			Total: plant.EnergyTotal,
			Unit:  "kWh",
		}, ttl)

		cache.Set("dashboard", plant, ttl)

		if mq != nil {
			log.Println("Polled ShineMonitor (optimized webQuery), pushing to MQTT...")
			mq.PublishState(mqtt.GlobalStatus{
				Status: &shinemonitor.StatusResponse{Status: statusStr},
				EnergyNow: &shinemonitor.EnergyNowResponse{
					TimeStamp:  time.Now(),
					Energy:     plant.OutputPower,
					Unit:       "kW",
					DeviceData: deviceData,
				},
				Summary: &shinemonitor.EnergySummaryResponse{
					Today: plant.Energy,
					Month: plant.EnergyMonth,
					Year:  plant.EnergyYear,
					Total: plant.EnergyTotal,
					Unit:  "kWh",
				},
				Dashboard:  plant,
				DeviceData: deviceData,
			})
		}
	} else {
		log.Printf("Poll resulted in incomplete fetching: err=%v", err)
		currentInverterState = StateOffline
	}
}
