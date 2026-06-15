package com.warmhouse.telemetry.controller;

import com.warmhouse.telemetry.model.TemperatureReading;
import com.warmhouse.telemetry.service.TemperatureService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

@RestController
@RequestMapping("/api/v1")
public class TemperatureController {

    private final TemperatureService temperatureService;

    public TemperatureController(TemperatureService temperatureService) {
        this.temperatureService = temperatureService;
    }

    @GetMapping("/health")
    public Map<String, String> health() {
        return Map.of("status", "ok", "service", "temperature-telemetry");
    }

    @GetMapping("/temperature/{deviceId}")
    public ResponseEntity<TemperatureReading> getTemperature(@PathVariable long deviceId) {
        return ResponseEntity.ok(temperatureService.getLatest(deviceId));
    }
}
