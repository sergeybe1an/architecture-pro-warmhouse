import json
import logging
import os
from datetime import datetime, timezone

import pika

logger = logging.getLogger(__name__)

EXCHANGE = "warmhouse.device.events"


def _connect():
    host = os.getenv("RABBITMQ_HOST", "rabbitmq")
    return pika.BlockingConnection(
        pika.ConnectionParameters(host=host, heartbeat=600, blocked_connection_timeout=300)
    )


ROUTING_KEYS = {
    "DeviceCreated": "device.created",
    "DeviceUpdated": "device.updated",
    "DeviceDeleted": "device.deleted",
}


def publish_event(event_type: str, sensor: dict) -> None:
    try:
        connection = _connect()
        channel = connection.channel()
        channel.exchange_declare(exchange=EXCHANGE, exchange_type="topic", durable=True)

        payload = {
            "eventType": event_type,
            "deviceId": sensor["id"],
            "type": sensor.get("type"),
            "location": sensor.get("location"),
            "status": sensor.get("status"),
            "occurredAt": datetime.now(timezone.utc).isoformat(),
        }

        channel.basic_publish(
            exchange=EXCHANGE,
            routing_key=ROUTING_KEYS[event_type],
            body=json.dumps(payload),
            properties=pika.BasicProperties(content_type="application/json", delivery_mode=2),
        )
        connection.close()
        logger.info("Published %s for device %s", event_type, sensor.get("id"))
    except Exception as exc:
        logger.warning("Failed to publish event %s: %s", event_type, exc)
