package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type ClerkEvent struct {
	EventType string `json:"event_type"`
	Data      struct {
		OrganizationID string `json:"organization_id"`
	} `json:"data"`
}

type TursoClient struct {
	APIKey string
}

func NewTursoClient(apiKey string) *TursoClient {
	return &TursoClient{APIKey: apiKey}
}

func (c *TursoClient) CreateDatabase(orgID string) error {
	url := fmt.Sprintf("https://api.turso.com/databases?org_id=%s", orgID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create database, status code: %d", resp.StatusCode)
	}

	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var event ClerkEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if event.EventType == "organization.created" {
		tursoClient := NewTursoClient(os.Getenv("TURSO_API_KEY"))
		if err := tursoClient.CreateDatabase(event.Data.OrganizationID); err != nil {
			http.Error(w, "Failed to create database", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/webhook", handler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
