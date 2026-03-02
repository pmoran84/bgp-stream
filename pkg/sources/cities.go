package sources

import (
	"encoding/json"
	"fmt"

	"github.com/sudorandom/bgp-stream/pkg/utils"
)

func DownloadWorldCities(dest string) error {
	return utils.DownloadFile(WorldCitiesURL, dest)
}

type CityDominance struct {
	Country             string
	Coordinates         []float64
	LogicalDominanceIPs float64 `json:"logical_dominance_ips"`
}

func FetchCityDominance() ([]CityDominance, error) {
	metaReader, err := utils.GetCachedReader(CityDominanceMetaURL, true, "[GEO-HUB-META]")
	if err != nil {
		return nil, err
	}
	defer func() { _ = metaReader.Close() }()

	var meta struct {
		MaxYear int `json:"max_year"`
	}
	if err := json.NewDecoder(metaReader).Decode(&meta); err != nil {
		return nil, err
	}

	url := fmt.Sprintf(CityDominanceDataURL, meta.MaxYear)
	dataReader, err := utils.GetCachedReader(url, true, "[GEO-HUB-DATA]")
	if err != nil {
		return nil, err
	}
	defer func() { _ = dataReader.Close() }()

	var cities []CityDominance
	if err := json.NewDecoder(dataReader).Decode(&cities); err != nil {
		return nil, err
	}

	return cities, nil
}
