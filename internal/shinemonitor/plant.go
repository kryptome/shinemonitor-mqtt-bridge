package shinemonitor

func (c *Client) GetPlantInfo() (*PlantInfoResponse, error) {
	plant, err := c.GetWebQueryPlants()
	if err != nil {
		return nil, err
	}

	return &PlantInfoResponse{
		InstallDate:        plant.Install,
		GridConnectionDate: plant.GTS,
		Address:            plant.Address,
		Profit: PlantProfit{
			TotalProfit: plant.Profit.UnitProfit, // Old API was empty, mapped to UnitProfit assuming closest metric available from dashboard
			TotalProfitValue: plant.Profit.Currency,
		},
	}, nil
}
