package shinemonitor

import (
	"encoding/json"
	"fmt"
	"strings"
)

// cleanTimelineJSON repairs the embedded Python structure dumped by the API
func cleanTimelineJSON(raw string, unit string) ([]byte, error) {
	str := strings.ReplaceAll(raw, "'", `"`)
	str = strings.ReplaceAll(str, `"val"`, `"Value"`)
	str = strings.ReplaceAll(str, `"ts"`, `"TimeStamp"`)
	str = strings.ReplaceAll(str, `}`, fmt.Sprintf(`, "Unit": "%s"}`, unit))
	return []byte(str), nil
}

func (c *Client) getTimelineData(action string, fieldName string, unit string) ([]EnergyTimelineResponse, error) {
	data, err := c.MakeRequest(action)
	if err != nil {
		return nil, err
	}

	dat, ok := data["dat"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing dat key in timeline")
	}

	rawString, ok := dat[fieldName].(string)
	if !ok || rawString == "" {
		return []EnergyTimelineResponse{}, nil // empty timeline is valid
	}

	cleanedJSON, err := cleanTimelineJSON(rawString, unit)
	if err != nil {
		return nil, err
	}

	type tempTimeline struct {
		TimeStamp string `json:"TimeStamp"`
		Value     string `json:"Value"`
		Unit      string `json:"Unit"`
	}

	var tmpList []tempTimeline
	if err := json.Unmarshal(cleanedJSON, &tmpList); err != nil {
		return nil, fmt.Errorf("failed to parse timeline JSON: %w", err)
	}

	var result []EnergyTimelineResponse
	for _, item := range tmpList {
		ts, _ := parseTimestamp(item.TimeStamp)
		result = append(result, EnergyTimelineResponse{
			TimeStamp: ts,
			Value:     item.Value,
			Unit:      item.Unit,
		})
	}

	return result, nil
}

// GetDeviceChart uses the exact chart query from user requested override for precise 'day' granularity
func (c *Client) GetDeviceChart(date string) ([]EnergyTimelineResponse, error) {
	sdate := fmt.Sprintf("%s 00:00:00", date)
	edate := fmt.Sprintf("%s 23:59:59", date)

	action := fmt.Sprintf("&action=queryDeviceChartFieldDetailData&devaddr=1&pn=%s&devcode=%s&sn=%s&field=output_power&precision=5&sdate=%s&edate=%s",
		c.Config.PN, c.Config.DevCode, c.Config.SN, sdate, edate)

	data, err := c.MakeRequest(action)
	if err != nil {
		return nil, err
	}

	jsonStr, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var res ChartDetailResponse
	if err := json.Unmarshal(jsonStr, &res); err != nil {
		return nil, err
	}

	if res.Err != 0 {
		return nil, fmt.Errorf("chart error code: %d", res.Err)
	}

	var result []EnergyTimelineResponse
	for _, item := range res.Dat {
		ts, _ := parseTimestamp(item.Key)
		result = append(result, EnergyTimelineResponse{
			TimeStamp: ts,
			Value:     item.Val,
			Unit:      "kW",
		})
	}

	return result, nil
}

func (c *Client) GetEnergyTimeline(date string) ([]EnergyTimelineResponse, error) {
	return c.GetDeviceChart(date)
}

func (c *Client) GetEnergyTimelineMonth(date string) ([]EnergyTimelineResponse, error) {
	action := fmt.Sprintf("&action=queryPlantEnergyMonthPerDay&plantid=%s&date=%s", c.Config.PlantID, date)
	return c.getTimelineData(action, "perday", "kWh")
}

func (c *Client) GetEnergyTimelineYear(date string) ([]EnergyTimelineResponse, error) {
	action := fmt.Sprintf("&action=queryPlantEnergyYearPerMonth&plantid=%s&date=%s", c.Config.PlantID, date)
	return c.getTimelineData(action, "permonth", "kWh")
}

func (c *Client) GetEnergyTimelineTotal() ([]EnergyTimelineResponse, error) {
	action := fmt.Sprintf("&action=queryPlantEnergyTotalPerYear&plantid=%s", c.Config.PlantID)
	return c.getTimelineData(action, "peryear", "kWh")
}
