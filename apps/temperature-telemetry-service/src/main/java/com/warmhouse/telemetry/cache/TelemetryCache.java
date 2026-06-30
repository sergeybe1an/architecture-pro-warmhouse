package com.warmhouse.telemetry.cache;

import com.warmhouse.telemetry.model.TemperatureReading;
import org.springframework.stereotype.Component;

import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;

@Component
public class TelemetryCache {
    private final Map<Long, TemperatureReading> cache = new ConcurrentHashMap<>();

    public Optional<TemperatureReading> get(long deviceId) {
        return Optional.ofNullable(cache.get(deviceId));
    }

    public void put(TemperatureReading reading) {
        cache.put(reading.getDeviceId(), reading);
    }

    public void remove(long deviceId) {
        cache.remove(deviceId);
    }
}
