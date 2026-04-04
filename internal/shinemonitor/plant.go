package shinemonitor

import (
	"encoding/json"
	"fmt"
)

func (c *Client) GetPlantInfo() (*PlantInfoResponse, error) {
	action := fmt.Sprintf("&action=queryPlantInfo&plantid=%s", c.Config.PlantID)
	data, err := c.MakeRequest(action)
	if err != nil {
		return &PlantInfoResponse{}, err
	}

	dat, ok := data["dat"].(map[string]interface{})
	if !ok {
		return &PlantInfoResponse{}, fmt.Errorf("missing dat key in plant info")
	}

	// Because dat is basically the shape of PlantInfoResponse, 
	// we can re-marshal it and unmarshal into our struct nicely.
	jsonBytes, err := json.Marshal(dat)
	if err != nil {
		return &PlantInfoResponse{}, err
	}

	var res PlantInfoResponse
	if err := json.Unmarshal(jsonBytes, &res); err != nil {
		return &PlantInfoResponse{}, err
	}

	return &res, nil
}
