package mqtt

import (
	"encoding/json"
	"fmt"
	"log"

	mqttclient "github.com/eclipse/paho.mqtt.golang"
	"github.com/kryptome/shinemonitor-mqtt-bridge/internal/config"
	"github.com/kryptome/shinemonitor-mqtt-bridge/internal/shinemonitor"
)

type Client struct {
	client mqttclient.Client
	cfg    *config.Config
}

func Connect(cfg *config.Config) (*Client, error) {
	opts := mqttclient.NewClientOptions()
	opts.AddBroker(cfg.MQTTBroker)
	opts.SetClientID(cfg.MQTTClientID)

	if cfg.MQTTUser != "" {
		opts.SetUsername(cfg.MQTTUser)
		opts.SetPassword(cfg.MQTTPassword)
	}

	stateTopic := fmt.Sprintf("%s/state", cfg.MQTTClientID)
	opts.SetWill(stateTopic, `{"status": "offline"}`, 1, true)

	opts.SetAutoReconnect(true)
	opts.SetOnConnectHandler(func(c mqttclient.Client) {
		log.Println("MQTT Connected!")
	})
	opts.SetConnectionLostHandler(func(c mqttclient.Client, err error) {
		log.Printf("MQTT Connection Lost: %v\n", err)
	})

	client := mqttclient.NewClient(opts)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return &Client{
		client: client,
		cfg:    cfg,
	}, nil
}

type HADiscoveryPayload struct {
	DeviceClass       string `json:"device_class,omitempty"`
	Name              string `json:"name"`
	StateTopic        string `json:"state_topic"`
	UnitOfMeasurement string `json:"unit_of_measurement,omitempty"`
	ValueTemplate     string `json:"value_template"`
	UniqueID          string `json:"unique_id"`
	Device            struct {
		Identifiers  []string `json:"identifiers"`
		Name         string   `json:"name"`
		Manufacturer string   `json:"manufacturer"`
		Model        string   `json:"model"`
	} `json:"device"`
}

func (m *Client) PublishDiscovery() {
	stateTopic := fmt.Sprintf("%s/state", m.cfg.MQTTClientID)

	sensors := []map[string]string{
		{"id": "status", "name": "Status", "class": "", "unit": "", "val": "{{ value_json.status.Status }}"},
		{"id": "energy_now", "name": "Current Power", "class": "power", "unit": "kW", "val": "{{ value_json.energy_now.Energy }}"},
		{"id": "energy_today", "name": "Energy Today", "class": "energy", "unit": "kWh", "val": "{{ value_json.summary.Today }}"},
		{"id": "energy_total", "name": "Energy Total", "class": "energy", "unit": "kWh", "val": "{{ value_json.summary.Total }}"},
		{"id": "efficiency", "name": "Power Efficiency", "class": "", "unit": "%", "val": "{{ value_json.dashboard.powerEfficiency }}"},
		{"id": "cf_value", "name": "CF Value", "class": "", "unit": "", "val": "{{ value_json.dashboard.cfValue }}"},
	}

	for _, s := range sensors {
		configTopic := fmt.Sprintf("homeassistant/sensor/shinemonitor_%s/%s/config", m.cfg.PN, s["id"])

		payload := HADiscoveryPayload{
			DeviceClass:       s["class"],
			Name:              fmt.Sprintf("ShineMonitor %s", s["name"]),
			StateTopic:        stateTopic,
			UnitOfMeasurement: s["unit"],
			ValueTemplate:     s["val"],
			UniqueID:          fmt.Sprintf("shinemonitor_%s_%s", m.cfg.PN, s["id"]),
		}
		payload.Device.Identifiers = []string{m.cfg.PN, m.cfg.SN}
		payload.Device.Name = fmt.Sprintf("ShineMonitor Inverter %s", m.cfg.SN)
		payload.Device.Manufacturer = "ShineMonitor"
		payload.Device.Model = "Inverter"

		payloadBytes, _ := json.Marshal(payload)
		token := m.client.Publish(configTopic, 1, true, string(payloadBytes))
		token.Wait()
	}

	log.Println("Published Home Assistant Discovery payload")
}

type GlobalStatus struct {
	Status    *shinemonitor.StatusResponse        `json:"status"`
	EnergyNow *shinemonitor.EnergyNowResponse     `json:"energy_now"`
	Summary   *shinemonitor.EnergySummaryResponse `json:"summary"`
	Dashboard *shinemonitor.WebQueryPlant         `json:"dashboard"`
}

func (m *Client) PublishState(state GlobalStatus) {
	stateTopic := fmt.Sprintf("%s/state", m.cfg.MQTTClientID)

	b, _ := json.Marshal(state)
	token := m.client.Publish(stateTopic, 1, true, string(b))
	token.Wait()
}
