package mapbuilder

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const filePath = "maps/custom.json"

func Load() (Map, error) {
	data, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		return Map{Nodes: []Node{}, Streets: []Street{}, Buildings: []Building{}, TrafficLights: []int{}}, nil
	}
	if err != nil {
		return Map{}, err
	}
	var result Map
	if err := json.Unmarshal(data, &result); err != nil {
		return Map{}, err
	}
	return result, result.Validate()
}

func Save(value Map) error {
	if err := value.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(filePath), ".custom-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	if _, err = temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err = temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporary.Name(), filePath)
}
