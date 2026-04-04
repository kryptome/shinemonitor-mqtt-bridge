package shinemonitor

import "time"

type StatusResponse struct {
	Status string `json:"Status"`
	Error  string `json:"Error,omitempty"`
}

type EnergyNowResponse struct {
	TimeStamp time.Time `json:"TimeStamp"`
	Energy    string    `json:"Energy"`
	Unit      string    `json:"Unit"`
	Error     string    `json:"Error,omitempty"`
}

type EnergySummaryResponse struct {
	Today  string `json:"Today"`
	Month  string `json:"Month"`
	Year   string `json:"Year"`
	Total  string `json:"Total"`
	Unit   string `json:"Unit"`
	Error  string `json:"Error,omitempty"`
}

type EnergyTimelineResponse struct {
	TimeStamp time.Time `json:"TimeStamp"`
	Value     string    `json:"Value"`
	Unit      string    `json:"Unit"`
}

type PlantInfoResponse struct {
	InstallDate        string       `json:"installDate"`
	GridConnectionDate string       `json:"gridConnectionDate"`
	Address            PlantAddress `json:"address"`
	Profit             PlantProfit  `json:"profit"`
}

type PlantAddress struct {
	Country string `json:"country"`
	City    string `json:"city"`
}

type PlantProfit struct {
	TotalProfit      string `json:"totalProfit,omitempty"`
	TotalProfitValue string `json:"totalProfitValue,omitempty"`
	UnitProfit       string `json:"unitProfit,omitempty"`
	Currency         string `json:"currency,omitempty"`
}

type WebQueryPlant struct {
	PID             int          `json:"pid"`
	Name            string       `json:"name"`
	Status          int          `json:"status"`
	OutputPower     string       `json:"outputPower"`
	Energy          string       `json:"energy"`
	EnergyMonth     string       `json:"energyMonth"`
	EnergyYear      string       `json:"energyYear"`
	EnergyTotal     string       `json:"energyTotal"`
	PowerEfficiency string       `json:"powerEfficiency"`
	CFValue         string       `json:"cfValue"`
	NominalPower    string       `json:"nominalPower"`
	LTS             string       `json:"lts,omitempty"`
	Install         string       `json:"install"`
	GTS             string       `json:"gts"`
	Address         PlantAddress `json:"address"`
	Profit          PlantProfit  `json:"profit"`
}

type WebQueryPlantsResponse struct {
	Dat struct {
		Total    int             `json:"total"`
		Plant    []WebQueryPlant `json:"plant"`
	} `json:"dat"`
	Err  int    `json:"err"`
	Desc string `json:"desc"`
}

type ChartDetailResponse struct {
	Dat []struct {
		Key string `json:"key"`
		Val string `json:"val"`
	} `json:"dat"`
	Err  int    `json:"err"`
	Desc string `json:"desc"`
}

type DeviceDataTitle struct {
	Title     string `json:"title"`
	Unit      string `json:"unit,omitempty"`
	IsDisplay int    `json:"isDisplay"`
}

type DeviceDataRow struct {
	Realtime bool     `json:"realtime"`
	Field    []string `json:"field"`
}

type DeviceDataOneDayPagingResponse struct {
	Err  int    `json:"err"`
	Desc string `json:"desc"`
	Dat  struct {
		Total    int               `json:"total"`
		Page     int               `json:"page"`
		Pagesize int               `json:"pagesize"`
		Title    []DeviceDataTitle `json:"title"`
		Row      []DeviceDataRow   `json:"row"`
	} `json:"dat"`
}
