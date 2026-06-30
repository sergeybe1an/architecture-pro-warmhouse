package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"smarthome/db"
	"smarthome/models"
	"smarthome/services"

	"github.com/gin-gonic/gin"
)

// SensorHandler handles sensor-related requests.
// In microservices mode it delegates device CRUD to Device Management Service
// and temperature enrichment to Temperature Telemetry Service (Strangler Fig).
type SensorHandler struct {
	DB                 *db.DB
	TemperatureService *services.TemperatureService
	DeviceClient       *services.DeviceClient
	TelemetryMSClient  *services.TelemetryMSClient
	UseMicroservices   bool
}

// NewSensorHandler creates a new SensorHandler.
func NewSensorHandler(
	db *db.DB,
	temperatureService *services.TemperatureService,
	deviceClient *services.DeviceClient,
	telemetryMSClient *services.TelemetryMSClient,
	useMicroservices bool,
) *SensorHandler {
	return &SensorHandler{
		DB:                 db,
		TemperatureService: temperatureService,
		DeviceClient:       deviceClient,
		TelemetryMSClient:  telemetryMSClient,
		UseMicroservices:   useMicroservices,
	}
}

// RegisterRoutes registers the sensor routes.
func (h *SensorHandler) RegisterRoutes(router *gin.RouterGroup) {
	sensors := router.Group("/sensors")
	{
		sensors.GET("", h.GetSensors)
		sensors.GET("/:id", h.GetSensorByID)
		sensors.POST("", h.CreateSensor)
		sensors.PUT("/:id", h.UpdateSensor)
		sensors.DELETE("/:id", h.DeleteSensor)
		sensors.PATCH("/:id/value", h.UpdateSensorValue)
		sensors.GET("/temperature/:location", h.GetTemperatureByLocation)
	}
}

func (h *SensorHandler) enrichTemperature(sensor *models.Sensor) {
	if sensor.Type != models.Temperature {
		return
	}

	sensorID := fmt.Sprintf("%d", sensor.ID)

	if h.UseMicroservices && h.TelemetryMSClient != nil {
		tempData, err := h.TelemetryMSClient.GetTemperatureByID(sensorID)
		if err == nil {
			sensor.Value = tempData.Value
			sensor.Status = tempData.Status
			sensor.LastUpdated = tempData.Timestamp
			if sensor.Unit == "" {
				sensor.Unit = tempData.Unit
			}
			log.Printf("Updated temperature for sensor %d via telemetry microservice", sensor.ID)
			return
		}
		log.Printf("Telemetry microservice failed for sensor %d, fallback to temperature-api: %v", sensor.ID, err)
	}

	tempData, err := h.TemperatureService.GetTemperatureByID(sensorID)
	if err == nil {
		sensor.Value = tempData.Value
		sensor.Status = tempData.Status
		sensor.LastUpdated = tempData.Timestamp
		log.Printf("Updated temperature for sensor %d via temperature-api", sensor.ID)
	} else {
		log.Printf("Failed to fetch temperature for sensor %d: %v", sensor.ID, err)
	}
}

// GetSensors handles GET /api/v1/sensors.
func (h *SensorHandler) GetSensors(c *gin.Context) {
	var sensors []models.Sensor
	var err error

	if h.UseMicroservices && h.DeviceClient != nil {
		sensors, err = h.DeviceClient.GetSensors()
	} else {
		sensors, err = h.DB.GetSensors(context.Background())
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i := range sensors {
		h.enrichTemperature(&sensors[i])
	}

	c.JSON(http.StatusOK, sensors)
}

// GetSensorByID handles GET /api/v1/sensors/:id.
func (h *SensorHandler) GetSensorByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID"})
		return
	}

	var sensor models.Sensor

	if h.UseMicroservices && h.DeviceClient != nil {
		sensor, err = h.DeviceClient.GetSensorByID(id)
	} else {
		sensor, err = h.DB.GetSensorByID(context.Background(), id)
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sensor not found"})
		return
	}

	h.enrichTemperature(&sensor)
	c.JSON(http.StatusOK, sensor)
}

// GetTemperatureByLocation handles GET /api/v1/sensors/temperature/:location.
func (h *SensorHandler) GetTemperatureByLocation(c *gin.Context) {
	location := c.Param("location")
	if location == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Location is required"})
		return
	}

	tempData, err := h.TemperatureService.GetTemperature(location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to fetch temperature data: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"location":    tempData.Location,
		"value":       tempData.Value,
		"unit":        tempData.Unit,
		"status":      tempData.Status,
		"timestamp":   tempData.Timestamp,
		"description": tempData.Description,
	})
}

// CreateSensor handles POST /api/v1/sensors.
func (h *SensorHandler) CreateSensor(c *gin.Context) {
	var sensorCreate models.SensorCreate
	if err := c.ShouldBindJSON(&sensorCreate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var sensor models.Sensor
	var err error

	if h.UseMicroservices && h.DeviceClient != nil {
		sensor, err = h.DeviceClient.CreateSensor(sensorCreate)
	} else {
		sensor, err = h.DB.CreateSensor(context.Background(), sensorCreate)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, sensor)
}

// UpdateSensor handles PUT /api/v1/sensors/:id (still in monolith DB — not yet migrated).
func (h *SensorHandler) UpdateSensor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID"})
		return
	}

	var sensorUpdate models.SensorUpdate
	if err := c.ShouldBindJSON(&sensorUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sensor, err := h.DB.UpdateSensor(context.Background(), id, sensorUpdate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sensor)
}

// DeleteSensor handles DELETE /api/v1/sensors/:id (still in monolith DB — not yet migrated).
func (h *SensorHandler) DeleteSensor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID"})
		return
	}

	err = h.DB.DeleteSensor(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sensor deleted successfully"})
}

// UpdateSensorValue handles PATCH /api/v1/sensors/:id/value (still in monolith DB).
func (h *SensorHandler) UpdateSensorValue(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID"})
		return
	}

	var request struct {
		Value  float64 `json:"value" binding:"required"`
		Status string  `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.DB.UpdateSensorValue(context.Background(), id, request.Value, request.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sensor value updated successfully"})
}
