package main

//Ai generated, idk GoLang yet, so expect some weird code. But it works
import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var rdb *redis.Client

// 1. INITIALIZE THE TRACER
func initTracer() *sdktrace.TracerProvider {
	// Point the exporter to our local Docker Jaeger instance
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint("localhost:4318"),
		otlptracehttp.WithInsecure(), // No TLS for local testing
	)
	if err != nil {
		log.Fatal(err)
	}

	// Label our service so it shows up beautifully in the Jaeger UI
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("go-ingress-router"),
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
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	log.Println("Connected to Redis Shield...")
}

func transferHandler(w http.ResponseWriter, r *http.Request) {
	// 2. THE BACKPACK: Grab the context that OpenTelemetry automatically created for this request
	ctx := r.Context()

	idemKey := r.Header.Get("X-Idempotency-Key")
	if idemKey == "" {
		http.Error(w, "Missing X-Idempotency-Key", http.StatusBadRequest)
		return
	}

	// Pass the backpack (ctx) to Redis! If this takes long, Jaeger will show it.
	val, err := rdb.Get(ctx, idemKey).Result()
	if err == nil {
		log.Printf("BLOCKED: Duplicate key: %s\n", idemKey)
		http.Error(w, fmt.Sprintf("Duplicate request: %s", val), http.StatusConflict)
		return
	}

	queryParams := r.URL.Query()
	springBootURL := "http://localhost:8080/api/transfer"

	// 3. THE HANDOFF: Use NewRequestWithContext so the Trace ID gets stuffed into the outgoing headers!
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, springBootURL, nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.URL.RawQuery = queryParams.Encode()

	// 4. THE INTERCEPTOR: Wrap the standard HTTP client with otelhttp so it actually transmits the trace
	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to reach Ledger Service", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		// Pass the backpack (ctx) to Redis again
		rdb.Set(ctx, idemKey, "COMPLETED", 24*time.Hour)
	}

	body, _ := io.ReadAll(resp.Body)
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func main() {
	tp := initTracer()
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// 5. WRAP THE ROUTER: This automatically starts a trace for every incoming request
	handler := http.HandlerFunc(transferHandler)
	wrappedHandler := otelhttp.NewHandler(handler, "Transfer_Endpoint")

	http.Handle("/api/transfer", wrappedHandler)

	fmt.Println("🚀 Go Router (with X-Ray) starting on http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
