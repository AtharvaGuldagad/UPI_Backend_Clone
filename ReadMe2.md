<SYSTEM_CONTEXT>
  <PROJECT_GOAL>
    Build a national-scale UPI (Unified Payments Interface) backend clone.
    System Requirements: Sustained high throughput, extreme low latency, absolute ACID data consistency, zero-trust architecture, robust double-spend prevention.
    Development Methodology: Tracer Bullet Approach (End-to-end vertical slices, learning incrementally).
  </PROJECT_GOAL>

  <TARGET_TECH_STACK>
    - Compute (Switch/Router Layer): Go (Golang)
    - Compute (Core Ledger/Business Logic): Java (Spring Boot)
    - Persistence (ACID Ledger): PostgreSQL (via Docker)
    - Cache & Shielding: Redis (via Docker)
    - Asynchronous/Event Broker: Apache Kafka
    - Edge/API Gateway: Kong or Apache APISIX
    - Observability: OpenTelemetry, Prometheus, Grafana, Jaeger
  </TARGET_TECH_STACK>

  <ARCHITECTURE>
    1. Gateway: Validates stateless JWTs at the edge.
    2. Go Router (The Ingress Switch): High-velocity reverse proxy. Checks Redis for idempotency keys to prevent double-spends. Routes valid requests to the Ledger.
    3. Spring Boot Ledger: The ACID core. Executes strict, transaction-bound database updates for money movement.
    4. Data Layer: PostgreSQL holds the state; Redis holds the idempotency cache.
    5. Event Layer (Kafka/Debezium): Implements the Outbox pattern to reliably publish ledger events without 2PC.
  </ARCHITECTURE>

  <PROGRESS_LOG>
    [COMPLETED] Tracer 1: The Core Ledger (Java/Spring Boot + PostgreSQL)
      - Bootstrapped Spring Boot application.
      - Created Account entity and Repository.
      - Implemented `@Transactional` TransferService to guarantee ACID account-to-account money movement.
      - Successfully connected to PostgreSQL via Docker and executed state changes.

    [COMPLETED] Tracer 2: The High-Velocity Ingress Switch (Golang)
      - Initialized a Go module.
      - Built a lightweight HTTP reverse proxy (`net/http`).
      - Successfully routed requests from port 8081 (Go) to 8080 (Java) and returned responses.

    [COMPLETED] Tracer 3: The Shield (Redis Idempotency)
      - Added Redis to docker-compose.
      - Upgraded the Go router to require an `X-Idempotency-Key` header.
      - Implemented Check-and-Set logic: Blocks duplicate UUIDs (HTTP 409 Conflict) and caches successful transaction keys with a 24-hour TTL.
      - Successfully tested and neutralized the Double-Spend problem.
  </PROGRESS_LOG>

  <NEXT_STEPS>
    [PENDING] Tracer 4: The Outbox (Kafka & Debezium)
      - Add an `outbox_events` table to the PostgreSQL database.
      - Update the Spring Boot `@Transactional` method to write the account update AND an event to the outbox table simultaneously.
      - Deploy Debezium to tail the Postgres WAL and push exactly-once events to an Apache Kafka topic.

    [PENDING] Tracer 5: The Perimeter (API Gateway & Auth)
      - Deploy Kong or APISIX in front of the Go router.
      - Configure JWT validation plugins.
      - Ensure the Go service is isolated and only accepts traffic from the Gateway.

    [PENDING] Tracer 6: Observability
      - Instrument Go and Java services with OpenTelemetry.
      - Push distributed traces to Jaeger to track a request from Gateway -> Go -> Java -> DB -> Kafka.
  </NEXT_STEPS>
</SYSTEM_CONTEXT>
