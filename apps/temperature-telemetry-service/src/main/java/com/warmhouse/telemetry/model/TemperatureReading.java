package com.warmhouse.telemetry.model;

import java.time.Instant;

public class TemperatureReading {
    private long deviceId;
    private double value;
    private String unit;
    private String location;
    private String status;
    private Instant recordedAt;
    private String source;

    public TemperatureReading() {
    }

    public TemperatureReading(long deviceId, double value, String unit, String location,
                              String status, Instant recordedAt, String source) {
        this.deviceId = deviceId;
        this.value = value;
        this.unit = unit;
        this.location = location;
        this.status = status;
        this.recordedAt = recordedAt;
        this.source = source;
    }

    public long getDeviceId() {
        return deviceId;
    }

    public void setDeviceId(long deviceId) {
        this.deviceId = deviceId;
    }

    public double getValue() {
        return value;
    }

    public void setValue(double value) {
        this.value = value;
    }

    public String getUnit() {
        return unit;
    }

    public void setUnit(String unit) {
        this.unit = unit;
    }

    public String getLocation() {
        return location;
    }

    public void setLocation(String location) {
        this.location = location;
    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public Instant getRecordedAt() {
        return recordedAt;
    }

    public void setRecordedAt(Instant recordedAt) {
        this.recordedAt = recordedAt;
    }

    public String getSource() {
        return source;
    }

    public void setSource(String source) {
        this.source = source;
    }
}
