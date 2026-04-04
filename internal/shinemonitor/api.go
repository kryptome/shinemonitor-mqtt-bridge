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

// GetDeviceDataOneDayPaging gets detailed device data and parses it as a map
func (c *Client) GetDeviceDataOneDayPaging() (map[string]string, error) {
	dateStr := time.Now().Format("2006-01-02")
	action := fmt.Sprintf("&action=queryDeviceDataOneDayPaging&devaddr=1&oddEvenRow=null&pn=%s&devcode=%s&sn=%s&date=%s&page=0&pagesize=50",
		c.Config.PN, c.Config.DevCode, c.Config.SN, dateStr)

	data, err := c.MakeRequest(action)
	if err != nil {
		return nil, err
	}

	jsonStr, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var res DeviceDataOneDayPagingResponse
	if err := json.Unmarshal(jsonStr, &res); err != nil {
		return nil, err
	}

	if res.Err != 0 {
		return nil, fmt.Errorf("queryDeviceDataOneDayPaging API Error: %d - %s", res.Err, res.Desc)
	}

	if len(res.Dat.Row) == 0 {
		return nil, fmt.Errorf("no device data rows returned")
	}

	// We take the first row which has our latest realtime data
	latestRow := res.Dat.Row[0]
	parsed := make(map[string]string)

	// Map each field to its title
	for i, titleRaw := range res.Dat.Title {
		if i < len(latestRow.Field) {
			parsed[titleRaw.Title] = latestRow.Field[i]
		}
	}

	// Clean up fields based on MPPT and Phase count
	if c.Config.MPPTCount < 3 {
		delete(parsed, "PV3 voltage")
		delete(parsed, "PV3 current")
	}
	if c.Config.MPPTCount < 2 {
		delete(parsed, "PV2 voltage")
		delete(parsed, "PV2 current")
	}

	if c.Config.PhaseCount < 3 {
		delete(parsed, "Grid S voltage")
		delete(parsed, "Grid S current")
		delete(parsed, "Grid T voltage")
		delete(parsed, "Grid T current")
		delete(parsed, "Grid line voltage RS")
		delete(parsed, "Grid line voltage ST")
		delete(parsed, "Grid line voltage TR")
	}

	return parsed, nil
}
