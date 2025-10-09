package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type ConfigKey struct {
	MailtrapUsername string `json:"mailtrap_username"`
	MailtrapPassword string `json:"mailtrap_password"`
}

func getSecret() ConfigKey {
	// 1. Read the file, needs to be in same /cmd/api folder
	file, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Failed to read config file for program - %v", err)
	}

	// 2. Unmarshal the JSON into the struct
	var configKey ConfigKey // Use the defined struct type
	err = json.Unmarshal(file, &configKey)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Optional: Print a confirmation message
	fmt.Printf("Email Config Loaded for user: %s\n", configKey.MailtrapUsername)

	// 3. Return the entire struct
	return configKey
}
