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
	StateClass        string `json:"state_class,omitempty"`
	Name              string `json:"name"`
	StateTopic        string `json:"state_topic"`
	UnitOfMeasurement string `json:"unit_of_measurement,omitempty"`
	ValueTemplate     string `json:"value_template"`
	UniqueID          string `json:"unique_id"`
	Device            struct {
		Identifiers      []string `json:"identifiers"`
		Name             string   `json:"name"`
		Manufacturer     string   `json:"manufacturer"`
		Model            string   `json:"model"`
		SerialNumber     string   `json:"serial_number,omitempty"`
		SWVersion        string   `json:"sw_version,omitempty"`
		ConfigurationURL string   `json:"configuration_url,omitempty"`
	} `json:"device"`
}

func (m *Client) PublishDiscovery() {
	stateTopic := fmt.Sprintf("%s/state", m.cfg.MQTTClientID)

	sensors := []map[string]string{
		{"id": "status", "name": "Solar Status", "class": "", "stateClass": "", "unit": "", "val": "{{ value_json.status.Status }}"},
		{"id": "current_power", "name": "Solar Current Power", "class": "power", "stateClass": "measurement", "unit": "kW", "val": "{{ value_json.energy_now.Energy }}"},
		{"id": "energy_today", "name": "Solar Energy Today", "class": "energy", "stateClass": "total_increasing", "unit": "kWh", "val": "{{ value_json.summary.Today }}"},
		{"id": "energy_total", "name": "Solar Energy Total", "class": "energy", "stateClass": "total_increasing", "unit": "kWh", "val": "{{ value_json.summary.Total }}"},
		{"id": "power_efficiency", "name": "Solar Power Efficiency", "class": "", "stateClass": "measurement", "unit": "%", "val": "{{ value_json.dashboard.powerEfficiency }}"},

		// AC power metrics
		{"id": "output_s", "name": "Output S (Apparent Power)", "class": "apparent_power", "stateClass": "measurement", "unit": "VA", "val": "{{ value_json.device_data['Output S'] }}"},
		{"id": "output_pf", "name": "Output Power Factor", "class": "power_factor", "stateClass": "measurement", "unit": "", "val": "{{ value_json.device_data['Output PF'] }}"},

		// PV1 (always shown)
		{"id": "pv1_voltage", "name": "PV1 Voltage", "class": "voltage", "stateClass": "measurement", "unit": "V", "val": "{{ value_json.device_data['PV1 voltage'] }}"},
		{"id": "pv1_current", "name": "PV1 Current", "class": "current", "stateClass": "measurement", "unit": "A", "val": "{{ value_json.device_data['PV1 current'] }}"},

		// Grid R (always shown)
		{"id": "grid_r_voltage", "name": "Grid R Voltage", "class": "voltage", "stateClass": "measurement", "unit": "V", "val": "{{ value_json.device_data['Grid R voltage'] }}"},
		{"id": "grid_r_current", "name": "Grid R Current", "class": "current", "stateClass": "measurement", "unit": "A", "val": "{{ value_json.device_data['Grid R current'] }}"},

		// Grid & performance
		{"id": "grid_frequency", "name": "Grid Frequency", "class": "frequency", "stateClass": "measurement", "unit": "Hz", "val": "{{ value_json.device_data['Grid frequency'] }}"},
		{"id": "bus_voltage", "name": "Bus Voltage", "class": "voltage", "stateClass": "measurement", "unit": "V", "val": "{{ value_json.device_data['bus voltage'] }}"},
		{"id": "cuf", "name": "CUF", "class": "", "stateClass": "measurement", "unit": "%", "val": "{{ value_json.device_data['CUF'] }}"},
		{"id": "pr", "name": "PR", "class": "", "stateClass": "measurement", "unit": "%", "val": "{{ value_json.device_data['PR'] }}"},
	}

	if m.cfg.MPPTCount >= 2 {
		sensors = append(sensors,
			map[string]string{"id": "pv2_voltage", "name": "PV2 Voltage", "class": "voltage", "stateClass": "measurement", "unit": "V", "val": "{{ value_json.device_data['PV2 voltage'] }}"},
			map[string]string{"id": "pv2_current", "name": "PV2 Current", "class": "current", "stateClass": "measurement", "unit": "A", "val": "{{ value_json.device_data['PV2 current'] }}"},
		)
	}

	if m.cfg.MPPTCount >= 3 {
		sensors = append(sensors,
			map[string]string{"id": "pv3_voltage", "name": "PV3 Voltage", "class": "voltage", "stateClass": "measurement", "unit": "V", "val": "{{ value_json.device_data['PV3 voltage'] }}"},
			map[string]string{"id": "pv3_current", "name": "PV3 Current", "class": "current", "stateClass": "measurement", "unit": "A", "val": "{{ value_json.device_data['PV3 current'] }}"},
		)
	}

	if m.cfg.PhaseCount >= 3 {
		sensors = append(sensors,
			map[string]string{"id": "grid_s_voltage", "name": "Grid S Voltage", "class": "voltage", "stateClass": "measurement", "unit": "V", "val": "{{ value_json.device_data['Grid S voltage'] }}"},
			map[string]string{"id": "grid_s_current", "name": "Grid S Current", "class": "current", "stateClass": "measurement", "unit": "A", "val": "{{ value_json.device_data['Grid S current'] }}"},
			map[string]string{"id": "grid_t_voltage", "name": "Grid T Voltage", "class": "voltage", "stateClass": "measurement", "unit": "V", "val": "{{ value_json.device_data['Grid T voltage'] }}"},
			map[string]string{"id": "grid_t_current", "name": "Grid T Current", "class": "current", "stateClass": "measurement", "unit": "A", "val": "{{ value_json.device_data['Grid T current'] }}"},
			map[string]string{"id": "grid_line_voltage_rs", "name": "Grid Line Voltage RS", "class": "voltage", "stateClass": "measurement", "unit": "V", "val": "{{ value_json.device_data['Grid line voltage RS'] }}"},
			map[string]string{"id": "grid_line_voltage_st", "name": "Grid Line Voltage ST", "class": "voltage", "stateClass": "measurement", "unit": "V", "val": "{{ value_json.device_data['Grid line voltage ST'] }}"},
			map[string]string{"id": "grid_line_voltage_tr", "name": "Grid Line Voltage TR", "class": "voltage", "stateClass": "measurement", "unit": "V", "val": "{{ value_json.device_data['Grid line voltage TR'] }}"},
		)
	}

	for _, s := range sensors {
		configTopic := fmt.Sprintf("homeassistant/sensor/solar_%s/config", s["id"])

		payload := HADiscoveryPayload{
			DeviceClass:       s["class"],
			StateClass:        s["stateClass"],
			Name:              s["name"],
			StateTopic:        stateTopic,
			UnitOfMeasurement: s["unit"],
			ValueTemplate:     s["val"],
			UniqueID:          fmt.Sprintf("solar_%s_%s", m.cfg.SN, s["id"]),
		}
		payload.Device.Identifiers = []string{m.cfg.SN}
		payload.Device.Name = "ShineMonitor Solar Inverter"
		payload.Device.Manufacturer = "Ksolare / ShineMonitor"
		payload.Device.Model = "Solar Inverter"
		payload.Device.SerialNumber = m.cfg.SN
		payload.Device.SWVersion = "1.0.0"
		payload.Device.ConfigurationURL = "http://localhost:8080/swagger/index.html"

		payloadBytes, _ := json.Marshal(payload)
		token := m.client.Publish(configTopic, 1, true, string(payloadBytes))
		token.Wait()
	}

	log.Println("Published Home Assistant Discovery payload")
}

type GlobalStatus struct {
	Status     *shinemonitor.StatusResponse        `json:"status"`
	EnergyNow  *shinemonitor.EnergyNowResponse     `json:"energy_now"`
	Summary    *shinemonitor.EnergySummaryResponse `json:"summary"`
	Dashboard  *shinemonitor.WebQueryPlant         `json:"dashboard"`
	DeviceData map[string]string                   `json:"device_data,omitempty"`
}

func (m *Client) PublishState(state GlobalStatus) {
	stateTopic := fmt.Sprintf("%s/state", m.cfg.MQTTClientID)

	b, _ := json.Marshal(state)
	token := m.client.Publish(stateTopic, 1, true, string(b))
	token.Wait()
}
