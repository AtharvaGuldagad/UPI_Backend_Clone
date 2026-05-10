# UPI Backend: Distributed Architecture & System Context
**Last Updated:** May 2026
**System Status:** Fully Instrumented, Event-Driven, Zero-Trust Architecture

## 1. Executive Summary
This document serves as the architectural source of truth for the UPI Backend Ledger system. We have built a high-throughput, idempotent, zero-trust financial transaction engine. The architecture spans across 6 different networked boundaries, handling cryptographic validation, distributed caching, ACID database transactions, Change Data Capture (CDC) streaming, and full-stack distributed tracing.

---

## 2. The Request Journey (The X-Ray View)
When a user executes a financial transfer, the request passes through the following pipeline:

1. **The Perimeter (Kong API Gateway):** Intercepts the HTTP request, mathematically verifies the HS256 JWT signature using a 32-byte secret, and forwards the authenticated request.
2. **The Switch (Go Ingress Router):** Receives the request, extracts the `X-Idempotency-Key` (UUID), and checks **Redis**. If the key exists, the request is immediately rejected to prevent double-spending. If new, the key is logged, and the request is passed forward with an OpenTelemetry W3C `traceparent` header.
3. **The Core (Java Spring Boot Ledger):** Accepts the request into its Tomcat worker threads. It opens a database connection via HikariCP and begins a `@Transactional` block.
4. **The Ledger (PostgreSQL):** Executes the ACID transaction:
   - Deducts balance from Account A (Row lock)
   - Adds balance to Account B (Row lock)
   - Inserts an event payload into the `outbox_event` table.
   - Executes the `Transaction.commit()` (triggering physical `fsync` to disk).
5. **The Wiretap (Debezium + Kafka):** Debezium silently monitors the Postgres Write-Ahead Log (WAL). When the transaction commits, it instantly streams the `outbox_event` payload into an Apache Kafka topic for downstream asynchronous processing.
6. **The Observer (Jaeger + OpenTelemetry):** Throughout steps 1-5, trace spans are generated and pushed via HTTP/Protobuf to Jaeger, visualizing the exact millisecond duration of every network hop and database query.

---

## 3. Core Architectural Decisions (ADRs)

### ADR 1: Zero-Trust at the Edge
* **Decision:** Do not rely on internal microservices to verify user identity. 
* **Implementation:** Kong API Gateway handles all JWT validation before traffic even hits our application logic. 
* **Learning:** Requires strict adherence to RFC 7518 (Keys must be >= 256 bits / 32 bytes for HS256).

### ADR 2: The Idempotency Shield
* **Decision:** Financial systems must never process the same request twice, even under heavy network retries.
* **Implementation:** A lightweight Go router sits in front of the Java Ledger, using Redis to store and check UUIDs.
* **Benefit:** Malicious or accidental duplicate requests are blocked in `< 3ms` without ever touching the slow SQL database or Java application.

### ADR 3: The Transactional Outbox Pattern
* **Problem:** The "Dual-Write" problem. We needed to update a bank balance AND publish a Kafka event reliably. If Kafka crashes, the money moves but the receipt is lost.
* **Implementation:** Java updates the balances and writes a row to an `outbox_event` table inside the *exact same* Postgres SQL transaction. Debezium reads the WAL file to guarantee at-least-once delivery to Kafka.

### ADR 4: Distributed Tracing over Logging
* **Problem:** In a distributed system, debugging by reading 5 different terminal logs is impossible.
* **Implementation:** Integrated OpenTelemetry. Go uses manual context propagation to stamp headers. Java uses the auto-instrumenting JVM agent (`opentelemetry-javaagent.jar`). 
* **Learning:** Protocol mismatches are fatal. The Java agent must be explicitly configured to speak `http/protobuf` on port `4318` to communicate with Jaeger correctly.

---

## 4. Performance & Bottleneck Analysis
During our initial `k6` stress testing (1,000 Virtual Users):
* **Baseline (No DB Logic):** ~2,600 Requests Per Second (RPS) at < 125ms latency.
* **Full Stack (ACID + Kafka + Trace):** ~400 RPS at 1.4s latency.

### The Bottleneck: "The Java Day Care"
Distributed tracing revealed that the `UPDATE` and `INSERT` SQL queries take nanoseconds in Postgres memory. The ultimate bottleneck is the `Transaction.commit()`. 
* **Why:** Postgres must perform an `fsync` to physically write the Write-Ahead Log to the hard disk to guarantee Durability (the 'D' in ACID).
* **The Chain Reaction:** Because disk writing takes ~9ms, Java's HikariCP database connections are held open longer. This exhausts Tomcat's worker threads, forcing incoming HTTP requests to wait in a queue, ultimately resulting in connection snaps (`EOF` errors) under severe load.

---

## 5. Technology Stack Summary
* **API Gateway:** Kong (Docker)
* **Ingress / Shield:** Go (net/http), Redis
* **Ledger Application:** Java 17, Spring Boot 3, Tomcat, Hibernate/JPA, HikariCP
* **Database:** PostgreSQL 15
* **Message Broker / CDC:** Apache Kafka, Zookeeper, Debezium Postgres Connector
* **Observability:** Jaeger, OpenTelemetry (Go SDK, Java Agent)
* **Testing:** k6 (JavaScript load testing)

## 6. Next Steps / Optimization Roadmap
1.  **Database Connection Tuning:** Optimize HikariCP pool sizes and Tomcat accept-counts to handle thread exhaustion.
2.  **Database Optimization:** Evaluate indexing or asynchronous commit strategies (balancing ACID vs. Throughput).
3.  **Kafka Consumer:** Build the downstream service that listens to the `bank.public.outbox_event` topic to send push notifications/receipts.
