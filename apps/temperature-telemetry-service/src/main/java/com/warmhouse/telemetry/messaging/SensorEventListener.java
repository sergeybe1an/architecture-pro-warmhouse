package com.warmhouse.telemetry.messaging;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.warmhouse.telemetry.service.TemperatureService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.amqp.rabbit.annotation.RabbitListener;
import org.springframework.stereotype.Component;

@Component
public class SensorEventListener {

    private static final Logger log = LoggerFactory.getLogger(SensorEventListener.class);

    private final TemperatureService temperatureService;
    private final ObjectMapper objectMapper;

    public SensorEventListener(TemperatureService temperatureService, ObjectMapper objectMapper) {
        this.temperatureService = temperatureService;
        this.objectMapper = objectMapper;
    }

    @RabbitListener(queues = "telemetry.sensor.events")
    public void handleDeviceEvent(String message) {
        try {
            JsonNode node = objectMapper.readTree(message);
            String eventType = node.path("eventType").asText();
            long deviceId = node.path("deviceId").asLong();

            log.info("Received event {} for device {}", eventType, deviceId);

            if ("DeviceDeleted".equals(eventType)) {
                temperatureService.invalidate(deviceId);
            } else {
                temperatureService.invalidate(deviceId);
            }
        } catch (Exception e) {
            log.warn("Failed to process device event: {}", e.getMessage());
        }
    }
}
