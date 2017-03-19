package main

// this file handles the config file - settings.json

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/oauth2"
)

var (
	configFilename = "./settings.json"
	configFilePerm = 0644
)

// Represents the config file
type Config struct {
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenExpiry  int64  `json:"tokenExpiry"`
	PlaylistID   string `json:"playlistID"`
	TokenType    string `json:"tokenType"`
}

// Validates the config file
func validateConfig(config Config) {
	if config.ClientID == "" {
		log.Fatal("Client ID is empty")
	}
	if config.ClientSecret == "" {
		log.Fatal("Client Secret is empty")
	}
}

// Reads the `settings.json` and returns data in struct `Config`
func readConfig() Config {
	raw, err := ioutil.ReadFile(configFilename)
	if err != nil {
		log.Fatal(err)
	}
	var config Config
	json.Unmarshal(raw, &config)
	validateConfig(config)
	return config
}

// Writes struct `Config` to the file `settings.json`
func writeConfig(config Config) {
	jsonBytes, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile(configFilename, jsonBytes, os.FileMode(configFilePerm))
}

// receives OAuth token and updates the config file accordingly
func updateConfig(token *oauth2.Token) {
	config := readConfig()
	config.AccessToken = token.AccessToken
	config.RefreshToken = token.RefreshToken
	config.TokenExpiry = token.Expiry.Unix()
	config.TokenType = token.TokenType
	writeConfig(config)
}
