# shinemonitor-mqtt-bridge

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8.svg)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![MQTT](https://img.shields.io/badge/MQTT-v5-2C7A2C.svg)](https://mqtt.org)
[![Home Assistant](https://img.shields.io/badge/Home%20Assistant-Compatible-%2312A3B6.svg)](https://www.home-assistant.io)

A simple Go bridge that pulls real-time solar data from ShineMonitor and publishes it to MQTT (with full Home Assistant auto-discovery) while also exposing a lightweight REST API.

I built this because I wanted a fast, single-binary solution that works great on a Raspberry Pi and doesn’t need any external services.

## Features

- Polls ShineMonitor (website scraping or official Open API)
- Full MQTT v5 support with Home Assistant MQTT Discovery
- REST API for manual checks and status
- Single binary — no extra broker or database needed
- Configurable via `.env`
- Runs anywhere Go runs (RPi, VPS, etc.)
- Automatic reconnects, QoS 1, and Last Will support

## How it works

1. Logs into ShineMonitor at regular intervals
2. Grabs your power station / inverter data
3. Cleans it up and publishes clean sensor values to MQTT
4. REST endpoints are available on port 8080

## Quick Start

```bash
git clone https://github.com/kryptome/shinemonitor-mqtt-bridge.git
cd shinemonitor-mqtt-bridge
cp .env.example .env
# Edit .env with your ShineMonitor login + MQTT details
go run main.go
```

## Configuration

Copy the example and fill in your details:

```bash
cp .env.example .env
```

All configuration is done through the `.env` file.

### Required ShineMonitor fields
- `SHINEMONITOR_USERNAME` – your username
- `SHINEMONITOR_PASSWORD` – your password
- `SHINEMONITOR_COMPANY_KEY` – company key (e.g. `bnrl_XXXXXXXXX`)
- `SHINEMONITOR_PLANT_ID` – plant ID (must be a string)
- `SHINEMONITOR_PN` – product number (Generally the Datalogger's SN)
- `SHINEMONITOR_SN` – serial number (Generally the Inverter's SN)
- `SHINEMONITOR_DEV_CODE` – device code

### MQTT settings
- `MQTT_BROKER` – broker address (e.g. `tcp://localhost:1883` or `tcp://192.168.1.100:1883`)
- `MQTT_USER` – username (leave empty if not required)
- `MQTT_PASSWORD` – password (leave empty if not required)
- `MQTT_CLIENT_ID` – client identifier (default: `shinemonitor-bridge`)

### General settings
- `POLL_INTERVAL` – how often to fetch data (e.g. `30s`, `1m`, `5m`)
- `LOG_LEVEL` – logging level (`debug`, `info`, `warn`, `error`)

## API Endpoints

The bridge exposes a lightweight REST API for querying real-time and historical data cleanly without hitting ShineMonitor directly for every query.
Once the bridge is running, you can explore the interactive API references via Swagger UI locally at: `http://localhost:8080/swagger/index.html`

- **`GET /status`** – Fetch online/offline status.
  - *Response:* `{"Status": "Online"}`
- **`GET /now`** – Get immediate live power generation.
  - *Response:* `{"TimeStamp": "2026-04-04T12:00:00Z", "Energy": "4.5", "Unit": "kW"}`
- **`GET /summary`** – Get summarized energy metrics (Today, Month, Year, Total).
  - *Response:* `{"Today": "12.3", "Month": "120.4", "Year": "1005.1", "Total": "3000", "Unit": "kWh"}`
- **`GET /dashboard`** – Advanced live diagnostics (Includes Efficiency, CF Value, output power).
- **`GET /timeline?date=2026-04-04`** – Array of intra-day charts recording historic production.
- **`GET /plant`** – Fetch deep configuration meta like install date and locations.

## Home Assistant Integration

Just enable the MQTT integration in Home Assistant. All sensors (PV power, battery SOC, grid export, etc.) will appear automatically.

## Project Status

- [ ] Shinemonitor login + data fetching
- [ ] MQTT publishing with HA discovery
- [ ] REST API
- [ ] Docker support
- [ ] Tests

## Contributing

Pull requests are very welcome! Especially helpful would be:
- Improved scraping or official API usage
- More sensor mappings
- Web UI for easier configuration
- Telegram or webhook notifications

## License

MIT © 2026 Piyush Mishra

---

Made for the solar + Home Assistant community.