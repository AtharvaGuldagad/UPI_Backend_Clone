package main

//Ai Generated, so expect weird code, but works.

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var rdb *redis.Client
var kafkaWriter *kafka.Writer

type TransferEvent struct {
	IdempotencyKey string `json:"idempotency_key"`
	FromAccount    string `json:"from_account"`
	ToAccount      string `json:"to_account"`
	Amount         string `json:"amount"`
}

func initTracer() *sdktrace.TracerProvider {
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint("localhost:4318"),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		log.Fatal(err)
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("go-ingress-producer"),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp
}

func init() {
	rdb = redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	kafkaWriter = &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "transfer-requests",
		Balancer: &kafka.LeastBytes{},
	}
	log.Println("Connected to Redis Shield and Kafka Queue...")
}

func transferHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idemKey := r.Header.Get("X-Idempotency-Key")
	if idemKey == "" {
		http.Error(w, "Missing X-Idempotency-Key", http.StatusBadRequest)
		return
	}

	// 1. Check Redis Shield
	val, err := rdb.Get(ctx, idemKey).Result()
	if err == nil {
		http.Error(w, fmt.Sprintf("Duplicate request: %s", val), http.StatusConflict)
		return
	}

	// 2. Build the Event Payload
	queryParams := r.URL.Query()
	event := TransferEvent{
		IdempotencyKey: idemKey,
		FromAccount:    queryParams.Get("from"),
		ToAccount:      queryParams.Get("to"),
		Amount:         queryParams.Get("amount"),
	}
	eventBytes, _ := json.Marshal(event)

	// 3. Extract the Trace Backpack for Kafka
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	var kafkaHeaders []kafka.Header
	for k, v := range carrier {
		kafkaHeaders = append(kafkaHeaders, kafka.Header{Key: k, Value: []byte(v)})
	}

	// 4. FIRE AND FORGET to Kafka
	msg := kafka.Message{
		Key:     []byte(idemKey), // Using IdemKey as Kafka Key guarantees order for the same transaction
		Value:   eventBytes,
		Headers: kafkaHeaders,
	}

	err = kafkaWriter.WriteMessages(ctx, msg)
	if err != nil {
		log.Printf("Failed to write to Kafka: %v", err)
		http.Error(w, "Broker unavailable", http.StatusServiceUnavailable)
		return
	}

	// 5. Lock it in Redis and Return FAST
	rdb.Set(ctx, idemKey, "PENDING_IN_KAFKA", 24*time.Hour)

	w.WriteHeader(http.StatusAccepted) // 202 Accepted
	w.Write([]byte(`{"status": "Transfer Pending", "message": "Your transaction is being processed."}`))
}

func main() {
	tp := initTracer()
	defer tp.Shutdown(context.Background())
	defer kafkaWriter.Close()

	handler := http.HandlerFunc(transferHandler)
	wrappedHandler := otelhttp.NewHandler(handler, "Produce_Transfer_Event")

	http.Handle("/api/transfer", wrappedHandler)

	fmt.Println("🚀 Go Kafka Producer starting on http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
