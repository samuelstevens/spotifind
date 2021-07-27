package core

import (
	"encoding/json"
  "io/ioutil"
)

type Config struct {
	Spotify struct {
		ClientId     string `json:"ClientId"`
		ClientSecret string `json:"ClientSecret"`
		RedirectUri  string `json:"RedirectUri"`
	} `json:"Spotify"`
}

func LoadConfig() (*Config, error) {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		return nil, err
	}

  var config Config
  err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

  return &config, nil
}
