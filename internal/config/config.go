package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Username     string
	Password     string
	CompanyKey   string
	PlantID      string
	PN           string
	SN           string
	DevCode      string
	MQTTBroker   string
	MQTTUser     string
	MQTTPassword string
	MQTTClientID string
	PollInterval time.Duration
	LogLevel     string
	Port         string
	MPPTCount    int
	PhaseCount   int
}

// LoadConfig reads configuration from environment variables, 
// using .env file as a fallback if present.
func LoadConfig() *Config {
	_ = godotenv.Load() // Ignore error if .env doesn't exist

	pollIntervalStr := os.Getenv("POLL_INTERVAL")
	if pollIntervalStr == "" {
		pollIntervalStr = "30s"
	}
	pollInterval, err := time.ParseDuration(pollIntervalStr)
	if err != nil {
		pollInterval = 30 * time.Second
	}

	clientID := os.Getenv("MQTT_CLIENT_ID")
	if clientID == "" {
		clientID = "shinemonitor-bridge"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mpptCountStr := os.Getenv("SHINEMONITOR_MPPT_COUNT")
	mpptCount, err := strconv.Atoi(mpptCountStr)
	if err != nil || mpptCount < 1 {
		mpptCount = 1
	}

	phaseCountStr := os.Getenv("SHINEMONITOR_PHASE_COUNT")
	phaseCount, err := strconv.Atoi(phaseCountStr)
	if err != nil || phaseCount < 1 {
		phaseCount = 1
	}

	return &Config{
		Username:     os.Getenv("SHINEMONITOR_USERNAME"),
		Password:     os.Getenv("SHINEMONITOR_PASSWORD"),
		CompanyKey:   os.Getenv("SHINEMONITOR_COMPANY_KEY"),
		PlantID:      os.Getenv("SHINEMONITOR_PLANT_ID"),
		PN:           os.Getenv("SHINEMONITOR_PN"),
		SN:           os.Getenv("SHINEMONITOR_SN"),
		DevCode:      os.Getenv("SHINEMONITOR_DEV_CODE"),
		MQTTBroker:   os.Getenv("MQTT_BROKER"),
		MQTTUser:     os.Getenv("MQTT_USER"),
		MQTTPassword: os.Getenv("MQTT_PASSWORD"),
		MQTTClientID: clientID,
		PollInterval: pollInterval,
		LogLevel:     os.Getenv("LOG_LEVEL"),
		Port:         port,
		MPPTCount:    mpptCount,
		PhaseCount:   phaseCount,
	}
}
