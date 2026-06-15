package com.warmhouse.telemetry.config;

import org.springframework.amqp.core.Binding;
import org.springframework.amqp.core.BindingBuilder;
import org.springframework.amqp.core.Queue;
import org.springframework.amqp.core.TopicExchange;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class RabbitConfig {

    public static final String EXCHANGE = "warmhouse.device.events";
    public static final String QUEUE = "telemetry.sensor.events";

    @Bean
    public TopicExchange deviceEventsExchange() {
        return new TopicExchange(EXCHANGE, true, false);
    }

    @Bean
    public Queue telemetryQueue() {
        return new Queue(QUEUE, true);
    }

    @Bean
    public Binding deviceCreatedBinding(Queue telemetryQueue, TopicExchange deviceEventsExchange) {
        return BindingBuilder.bind(telemetryQueue).to(deviceEventsExchange).with("device.#");
    }
}
