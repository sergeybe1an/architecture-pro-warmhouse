package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TelemetryMSClient calls Temperature Telemetry Service (Java).
type TelemetryMSClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewTelemetryMSClient(baseURL string) *TelemetryMSClient {
	return &TelemetryMSClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type telemetryReading struct {
	DeviceID   int64     `json:"deviceId"`
	Value      float64   `json:"value"`
	Unit       string    `json:"unit"`
	Location   string    `json:"location"`
	Status     string    `json:"status"`
	RecordedAt time.Time `json:"recordedAt"`
	Source     string    `json:"source"`
}

func (c *TelemetryMSClient) GetTemperatureByID(sensorID string) (*TemperatureResponse, error) {
	url := fmt.Sprintf("%s/api/v1/temperature/%s", c.BaseURL, sensorID)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("telemetry service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("telemetry service returned status %d", resp.StatusCode)
	}

	var reading telemetryReading
	if err := json.NewDecoder(resp.Body).Decode(&reading); err != nil {
		return nil, fmt.Errorf("decode telemetry service response: %w", err)
	}

	return &TemperatureResponse{
		Value:       reading.Value,
		Unit:        reading.Unit,
		Timestamp:   reading.RecordedAt,
		Location:    reading.Location,
		Status:      reading.Status,
		SensorID:    sensorID,
		SensorType:  "temperature",
		Description: "Temperature from telemetry service",
	}, nil
}
