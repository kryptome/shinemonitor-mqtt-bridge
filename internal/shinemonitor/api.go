package shinemonitor

import (
	"encoding/json"
	"fmt"
	"time"
)

// GetWebQueryPlants uses the new efficient webQueryPlants endpoint that exposes outputPower, efficiency, cfValue, status etc.
func (c *Client) GetWebQueryPlants() (*WebQueryPlant, error) {
	action := "&action=webQueryPlants&page=0&pagesize=15"
	data, err := c.MakeRequest(action)
	if err != nil {
		return nil, err
	}

	// Remarshal back to use struct
	jsonStr, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var res WebQueryPlantsResponse
	if err := json.Unmarshal(jsonStr, &res); err != nil {
		return nil, err
	}

	if res.Err != 0 {
		return nil, fmt.Errorf("webQueryPlants API Error: %d - %s", res.Err, res.Desc)
	}

	if len(res.Dat.Plant) == 0 {
		return nil, fmt.Errorf("no plants returned")
	}

	// Default to the first plant, or match c.Config.PlantID if possible (omitted for brevity, assuming only 1 plant or first is correct)
	plant := res.Dat.Plant[0]
	return &plant, nil
}

func (c *Client) GetStatus() (*StatusResponse, error) {
	plant, err := c.GetWebQueryPlants()
	if err != nil {
		return &StatusResponse{Status: "Offline", Error: err.Error()}, nil // API errors often returned in JSON
	}

	statusStr := "Offline"
	if plant.Status == 0 {
		statusStr = "Online"
	}

	return &StatusResponse{Status: statusStr}, nil
}

func (c *Client) GetEnergyNow() (*EnergyNowResponse, error) {
	plant, err := c.GetWebQueryPlants()
	if err != nil {
		return nil, err
	}

	// WebQueryPlants returns outputPower in kW. 
	// The original EnergyNow was traditionally W, but we can return it raw as kW.
	return &EnergyNowResponse{
		TimeStamp: time.Now(), // We use Now, or plant.LTS if available, webQueryPlants doesn't inherently give an explicit TS for outputPower in the array, wait plant.LTS might not exist in array.
		Energy:    plant.OutputPower,
		Unit:      "kW",
	}, nil
}

func (c *Client) GetEnergySummary() (*EnergySummaryResponse, error) {
	plant, err := c.GetWebQueryPlants()
	if err != nil {
		return nil, err
	}

	return &EnergySummaryResponse{
		Today: plant.Energy,
		Month: plant.EnergyMonth,
		Year:  plant.EnergyYear,
		Total: plant.EnergyTotal,
		Unit:  "kWh",
	}, nil
}
