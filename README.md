# shinemonitor-mqtt-bridge

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8.svg)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![MQTT](https://img.shields.io/badge/MQTT-v5-2C7A2C.svg)](https://mqtt.org)
[![Home Assistant](https://img.shields.io/badge/Home%20Assistant-Compatible-%2312A3B6.svg)](https://www.home-assistant.io)

A simple Go bridge that pulls real-time solar data from ShineMonitor and publishes it to MQTT (with full Home Assistant auto-discovery) while also exposing a lightweight REST API.

I built this because I wanted a fast, single-binary solution that works great on a Raspberry Pi and doesn't need any external services.

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

## Installation

### Option 1: One-command Install (Linux / Raspberry Pi)

The easiest way. This builds the binary, installs it as a system service, and starts it automatically:

```bash
git clone https://github.com/kryptome/shinemonitor-mqtt-bridge.git
cd shinemonitor-mqtt-bridge
cp .env.example .env
# Edit .env with your credentials
bash install.sh
```

That's it — the bridge will start immediately and auto-restart on boot.

### Option 2: Pre-compiled Binaries

Pre-compiled binaries for Linux (amd64, arm64) and macOS will be available on the [GitHub Releases](https://github.com/kryptome/shinemonitor-mqtt-bridge/releases) page. Just download, extract, and run.

### Option 3: Build from Source

If you have Go installed:

```bash
git clone https://github.com/kryptome/shinemonitor-mqtt-bridge.git
cd shinemonitor-mqtt-bridge
go build -o shinemonitor-mqtt-bridge main.go
./shinemonitor-mqtt-bridge
```

## Running as a Service (Recommended)

For a real setup (especially on a Raspberry Pi), you want the bridge to start on boot and restart if it crashes.

A ready-to-use systemd service file is included: `shinemonitor-mqtt-bridge.service`

### Manual systemd setup

```bash
# Build the binary
go build -o shinemonitor-mqtt-bridge main.go

# Copy binary + config to /opt
sudo mkdir -p /opt/shinemonitor-mqtt-bridge
sudo cp shinemonitor-mqtt-bridge /opt/shinemonitor-mqtt-bridge/
sudo cp .env /opt/shinemonitor-mqtt-bridge/

# Install the service + logrotate
sudo cp shinemonitor-mqtt-bridge.service /etc/systemd/system/
sudo cp shinemonitor-mqtt-bridge.logrotate /etc/logrotate.d/shinemonitor-mqtt-bridge
sudo systemctl daemon-reload
sudo systemctl enable --now shinemonitor-mqtt-bridge
```

### Logs

Logs are written to `/var/log/shinemonitor-mqtt-bridge.log` and are **automatically rotated daily** (7 days retained, compressed). No manual cleanup needed.

### Useful commands

```bash
sudo systemctl status shinemonitor-mqtt-bridge    # Check status
tail -f /var/log/shinemonitor-mqtt-bridge.log      # Follow logs
sudo systemctl restart shinemonitor-mqtt-bridge    # Restart
```

> **Note:** The service file assumes install path `/opt/shinemonitor-mqtt-bridge`. Edit the `.service` file if your setup is different.

## Docker

If you prefer Docker, a `Dockerfile` and `docker-compose.yml` are included.

### Using Docker Compose (recommended)

```bash
cp .env.example .env
# Edit .env with your credentials
docker compose up -d
```

### Using Docker directly

```bash
docker build -t shinemonitor-mqtt-bridge .
docker run -d --name shinemonitor-bridge --env-file .env -p 8080:8080 shinemonitor-mqtt-bridge
```

> **Tip:** If your MQTT broker runs on the host machine, uncomment the `extra_hosts` section in `docker-compose.yml` and use `host.docker.internal` as the broker address.

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

### Advanced Configuration
By default, the bridge exposes sensors for a single MPPT and single phase (PV1 + Grid R). This keeps your Home Assistant integration clean and noise-free. 
If your inverter uses multiple MPPTs or is a 3-Phase inverter, you can individually enable the extra sensors by adding these *optional* keys to your `.env`:
- `SHINEMONITOR_MPPT_COUNT` – set to `2` or `3` to enable PV2 and PV3 sensors. Defaults to `1`.
- `SHINEMONITOR_PHASE_COUNT` – set to `3` to automatically enable Grid S, Grid T, and Line Voltages (RS, ST, TR). Defaults to `1`.

Single-phase / single-MPPT users can completely skip adding these keys.

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

Just enable the MQTT integration in Home Assistant. All sensors will appear automatically.

### Exposed Sensors
The bridge pulls data from both the dashboard and detailed query endpoints to expose:
- **Global Status**: Online Status, Current Power, Power Efficiency, CF Value
- **Energy Metrics**: Energy Today, Energy Total, Instrument Power
- **Performance**: CUF (%), PR (%)
- **Voltages & Currents**:
  - PV1, PV2, PV3 Voltage and Current
  - Grid R, S, T Voltage and Current
  - Grid Line Voltages (RS, ST, TR)
- **AC Metrics**: Output Power, Output S (VA), Output Q (VAr), Output PF, Grid Frequency, Bus Voltage
- **Inverter Diagnostics**: Waiting Time, ISO, DCI, GFCI

## Project Status

- [x] Shinemonitor login + data fetching
- [x] MQTT publishing with HA discovery
- [x] REST API
- [x] Docker support
- [ ] Tests
- [ ] GitHub Release binaries

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