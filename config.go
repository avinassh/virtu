package virtu

// this file handles the config file - settings.json

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

var (
	configFilename = "./settings.json"
	configFilePerm = 0644
)

// Represents the config file
type Config struct {
	ClintID      string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	PlaylistID   string `json:"playlistID"`
}

// Reads the `settings.json` and returns data in struct `Config`
func readConfig() Config {
	raw, err := ioutil.ReadFile(configFilename)
	if err != nil {
		log.Fatal(err)
	}
	var config Config
	json.Unmarshal(raw, &config)
	return config
}

// Writes struct `Config` to the file `settings.json`
func writeConfig(config Config) {
	jsonBytes, err := json.Marshal(config)
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile(configFilename, jsonBytes, os.FileMode(configFilePerm))
}
