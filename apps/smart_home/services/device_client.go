package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"smarthome/models"
)

// DeviceClient calls Device Management Service (Python).
type DeviceClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewDeviceClient(baseURL string) *DeviceClient {
	return &DeviceClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type deviceSensor struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	Location  string     `json:"location"`
	Unit      *string    `json:"unit"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func toSensor(d deviceSensor) models.Sensor {
	unit := ""
	if d.Unit != nil {
		unit = *d.Unit
	}
	return models.Sensor{
		ID:          d.ID,
		Name:        d.Name,
		Type:        models.SensorType(d.Type),
		Location:    d.Location,
		Unit:        unit,
		Status:      d.Status,
		LastUpdated: d.UpdatedAt,
		CreatedAt:   d.CreatedAt,
	}
}

func (c *DeviceClient) GetSensors() ([]models.Sensor, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/v1/sensors")
	if err != nil {
		return nil, fmt.Errorf("device service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device service returned status %d", resp.StatusCode)
	}

	var payload []deviceSensor
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode device service response: %w", err)
	}

	sensors := make([]models.Sensor, len(payload))
	for i, d := range payload {
		sensors[i] = toSensor(d)
	}
	return sensors, nil
}

func (c *DeviceClient) GetSensorByID(id int) (models.Sensor, error) {
	resp, err := c.HTTPClient.Get(fmt.Sprintf("%s/api/v1/sensors/%d", c.BaseURL, id))
	if err != nil {
		return models.Sensor{}, fmt.Errorf("device service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return models.Sensor{}, fmt.Errorf("sensor not found")
	}
	if resp.StatusCode != http.StatusOK {
		return models.Sensor{}, fmt.Errorf("device service returned status %d", resp.StatusCode)
	}

	var payload deviceSensor
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return models.Sensor{}, fmt.Errorf("decode device service response: %w", err)
	}
	return toSensor(payload), nil
}

func (c *DeviceClient) CreateSensor(sensorCreate models.SensorCreate) (models.Sensor, error) {
	body, err := json.Marshal(sensorCreate)
	if err != nil {
		return models.Sensor{}, err
	}

	resp, err := c.HTTPClient.Post(c.BaseURL+"/api/v1/sensors", "application/json", bytes.NewReader(body))
	if err != nil {
		return models.Sensor{}, fmt.Errorf("device service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return models.Sensor{}, fmt.Errorf("device service returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var payload deviceSensor
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return models.Sensor{}, fmt.Errorf("decode device service response: %w", err)
	}
	return toSensor(payload), nil
}
