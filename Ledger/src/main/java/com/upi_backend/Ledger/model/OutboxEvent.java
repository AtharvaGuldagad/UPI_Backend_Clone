package com.upi_backend.Ledger.model;

import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.Column;
import java.time.Instant;
import java.util.UUID;

@Entity
public class OutboxEvent {
    @Id
    private String eventId;
    private String aggregateType;
    private String eventType;     
    
    @Column(columnDefinition = "TEXT")
    private String payload;       // JSON message
    
    private Instant createdAt;

    public OutboxEvent() {}

    public OutboxEvent(String aggregateType, String eventType, String payload) {
        this.eventId = UUID.randomUUID().toString();
        this.aggregateType = aggregateType;
        this.eventType = eventType;
        this.payload = payload;
        this.createdAt = Instant.now();
    }

    // Getters
    public String getEventId() { return eventId; }
    public String getAggregateType() { return aggregateType; }
    public String getEventType() { return eventType; }
    public String getPayload() { return payload; }
    public Instant getCreatedAt() { return createdAt; }
}