import os
from contextlib import contextmanager
from datetime import datetime, timezone
from typing import Optional

import psycopg2
import psycopg2.extras
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

from events import publish_event

DATABASE_URL = os.getenv(
    "DATABASE_URL",
    "postgres://postgres:postgres@postgres:5432/device_management",
)

app = FastAPI(title="Device Management Service", version="1.0.0")


class SensorCreate(BaseModel):
    name: str
    type: str
    location: str
    unit: Optional[str] = None


class SensorUpdate(BaseModel):
    name: Optional[str] = None
    type: Optional[str] = None
    location: Optional[str] = None
    unit: Optional[str] = None
    status: Optional[str] = None


class Sensor(BaseModel):
    id: int
    name: str
    type: str
    location: str
    unit: Optional[str] = None
    status: str
    created_at: datetime
    updated_at: datetime


@contextmanager
def get_connection():
    conn = psycopg2.connect(DATABASE_URL)
    try:
        yield conn
        conn.commit()
    except Exception:
        conn.rollback()
        raise
    finally:
        conn.close()


def row_to_sensor(row: dict) -> Sensor:
    return Sensor(
        id=row["id"],
        name=row["name"],
        type=row["type"],
        location=row["location"],
        unit=row.get("unit"),
        status=row["status"],
        created_at=row["created_at"],
        updated_at=row["updated_at"],
    )


@app.get("/health")
def health():
    return {"status": "ok", "service": "device-management"}


@app.get("/api/v1/sensors", response_model=list[Sensor])
def list_sensors():
    with get_connection() as conn:
        with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
            cur.execute(
                "SELECT id, name, type, location, unit, status, created_at, updated_at "
                "FROM sensors ORDER BY id"
            )
            return [row_to_sensor(row) for row in cur.fetchall()]


@app.get("/api/v1/sensors/{sensor_id}", response_model=Sensor)
def get_sensor(sensor_id: int):
    with get_connection() as conn:
        with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
            cur.execute(
                "SELECT id, name, type, location, unit, status, created_at, updated_at "
                "FROM sensors WHERE id = %s",
                (sensor_id,),
            )
            row = cur.fetchone()
            if not row:
                raise HTTPException(status_code=404, detail="Sensor not found")
            return row_to_sensor(row)


@app.post("/api/v1/sensors", response_model=Sensor, status_code=201)
def create_sensor(payload: SensorCreate):
    now = datetime.now(timezone.utc)
    with get_connection() as conn:
        with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
            cur.execute(
                """
                INSERT INTO sensors (name, type, location, unit, status, created_at, updated_at)
                VALUES (%s, %s, %s, %s, 'inactive', %s, %s)
                RETURNING id, name, type, location, unit, status, created_at, updated_at
                """,
                (payload.name, payload.type, payload.location, payload.unit, now, now),
            )
            sensor = row_to_sensor(cur.fetchone())

    publish_event("DeviceCreated", sensor.model_dump())
    return sensor


@app.put("/api/v1/sensors/{sensor_id}", response_model=Sensor)
def update_sensor(sensor_id: int, payload: SensorUpdate):
    fields = []
    values = []
    for field in ("name", "type", "location", "unit", "status"):
        value = getattr(payload, field)
        if value is not None:
            fields.append(f"{field} = %s")
            values.append(value)

    if not fields:
        raise HTTPException(status_code=400, detail="No fields to update")

    fields.append("updated_at = %s")
    values.append(datetime.now(timezone.utc))
    values.append(sensor_id)

    with get_connection() as conn:
        with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
            cur.execute(
                f"""
                UPDATE sensors SET {", ".join(fields)}
                WHERE id = %s
                RETURNING id, name, type, location, unit, status, created_at, updated_at
                """,
                values,
            )
            row = cur.fetchone()
            if not row:
                raise HTTPException(status_code=404, detail="Sensor not found")
            sensor = row_to_sensor(row)

    publish_event("DeviceUpdated", sensor.model_dump())
    return sensor


@app.delete("/api/v1/sensors/{sensor_id}")
def delete_sensor(sensor_id: int):
    with get_connection() as conn:
        with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
            cur.execute(
                "SELECT id, name, type, location, unit, status, created_at, updated_at "
                "FROM sensors WHERE id = %s",
                (sensor_id,),
            )
            row = cur.fetchone()
            if not row:
                raise HTTPException(status_code=404, detail="Sensor not found")

            cur.execute("DELETE FROM sensors WHERE id = %s", (sensor_id,))

    publish_event("DeviceDeleted", {"id": sensor_id, "type": row["type"], "location": row["location"], "status": row["status"]})
    return {"message": "Sensor deleted successfully"}
