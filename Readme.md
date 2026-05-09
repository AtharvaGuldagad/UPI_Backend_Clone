# UPI Backend Clone

I'm building a national-scale UPI (Unified Payments Interface) backend clone. My previous experience is in building high-traffic flash sale architectures—handling massive, bursty traffic (like 5000 requests in 10 seconds) relying on eventual consistency. This project is my transition into the deep end: sustained, unrelenting workloads that demand *absolute* strong consistency.

## The Goals
This is a forcing function to level up my system design and engineering skills. Here is exactly what I am setting out to achieve:

* **Scale to Enterprise Traffic:** Build a heavy, enterprise-level backend capable of sustaining 7,500 transactions per second (TPS).
* **Master Zero-Trust Security:** Learn how to architect a genuinely secure, bulletproof financial system from the ground up, starting at the API Gateway.
* **Push it to the Breaking Point:** Load test this entire architecture to its absolute limits to see exactly where and how it fails XD.
* **Solve Distributed Observability:** Implement comprehensive distributed tracing (OpenTelemetry/Jaeger). Debugging a failed transaction across five different microservices at 7,500 TPS without tracing is suicide.
* **Nail Disaster Recovery:** Design a failover mechanism where a node can die mid-transaction without dropping the request, double-charging the user, or permanently locking their funds.

## How I'm Building This: Tracer Bullets
Because this system is massively complex, I'm not building it layer by layer. I am using a tracer bullet methodology. I need to prove the hardest part of the architecture works from front to back before I build anything else.

**Phase 1 (Current Focus): The Core Ledger Service**
Right now, the only thing that matters is the ledger. I am building the core Java/Spring Boot ledger backed by PostgreSQL. If the system cannot handle concurrent, append-only writes to a single ledger row without row-lock deadlocks or double-spending, the rest of the microservices don't matter. 

Once the core ledger is bulletproof and correctly utilizing the Outbox Pattern for atomic database writes, I will start expanding outward to the Go-based router layer and Kafka event broker. 

## The Big Problems I'm Solving Here
Financial systems are unforgiving. These are the architectural hurdles I'm focusing on:
* **The Double Spend Problem:** Guaranteeing exactly-once processing using idempotency keys (UUID caching) despite inevitable network retries and failures.
* **Database Contention:** Handling highly concurrent writes to a single account without grinding the database to a halt with deadlocks.
* **Distributed Transactions:** Orchestrating atomic rollbacks across isolated bank microservices using the Saga Pattern instead of relying on slow, blocking Two-Phase Commits (2PC).
* **Stateless Auth:** Executing stateless JWT authentication at the API Gateway with a distributed in-memory deny list so we don't bloat latency on every single hop.

## The Target Tech Stack
This is where we are heading as the tracer bullets expand:
* **Compute (Switch/Router):** Go (Golang)
* **Compute (Core Ledger):** Java (Spring Boot)
* **Persistence (ACID Ledger):** PostgreSQL (Evaluating CockroachDB for horizontal SQL scaling later)
* **Event Broker:** Apache Kafka
* **Cache & Deduplication:** Redis Cluster
* **API Gateway:** Kong or Apache APISIX
* **Observability:** OpenTelemetry, Prometheus, Grafana, Jaeger