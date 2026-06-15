package com.warmhouse.telemetry.service;

import com.warmhouse.telemetry.cache.TelemetryCache;
import com.warmhouse.telemetry.model.TemperatureReading;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;

import java.time.Instant;
import java.util.Map;

@Service
public class TemperatureService {

    private final RestTemplate restTemplate;
    private final TelemetryCache cache;
    private final String temperatureApiUrl;

    public TemperatureService(RestTemplate restTemplate,
                              TelemetryCache cache,
                              @Value("${temperature-api.url}") String temperatureApiUrl) {
        this.restTemplate = restTemplate;
        this.cache = cache;
        this.temperatureApiUrl = temperatureApiUrl;
    }

    public TemperatureReading getLatest(long deviceId) {
        return cache.get(deviceId).orElseGet(() -> fetchAndCache(deviceId));
    }

    @SuppressWarnings("unchecked")
    private TemperatureReading fetchAndCache(long deviceId) {
        String url = temperatureApiUrl + "/temperature/" + deviceId;
        ResponseEntity<Map> response = restTemplate.getForEntity(url, Map.class);
        Map<String, Object> body = response.getBody();

        if (body == null) {
            throw new IllegalStateException("Empty response from temperature-api");
        }

        double value = ((Number) body.get("value")).doubleValue();
        String unit = String.valueOf(body.get("unit"));
        String location = String.valueOf(body.get("location"));
        String status = String.valueOf(body.get("status"));

        TemperatureReading reading = new TemperatureReading(
                deviceId,
                value,
                unit,
                location,
                status,
                Instant.now(),
                "temperature-api"
        );
        cache.put(reading);
        return reading;
    }

    public void invalidate(long deviceId) {
        cache.remove(deviceId);
    }
}
