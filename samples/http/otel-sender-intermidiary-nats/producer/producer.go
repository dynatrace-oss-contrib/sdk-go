package main

import (
	"context"
	"fmt"
	"html"
	"log"
	"net/http"
	"time"

	otelObs "github.com/cloudevents/sdk-go/observability/opentelemetry/v2/client"
	"github.com/cloudevents/sdk-go/samples/http/otel-sender-intermidiary-nats/instrumentation"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/google/uuid"
)

const (
	serviceName = "producer"
)

func main() {
	shutdown := instrumentation.InitOTelSdk(serviceName)
	defer shutdown()

	http.HandleFunc("/send", sendEvent)

	log.Println("Producer listening on localhost:8081")

	log.Fatal(http.ListenAndServe(":8081", nil))
}

func sendEvent(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))

	event := cloudevents.NewEvent()
	event.SetSource("example/uri")
	event.SetType("example.type")
	event.SetID(uuid.New().String())
	event.SetTime(time.Now().UTC())
	event.SetData(cloudevents.ApplicationJSON, map[string]string{"hello": "world"})

	// create the cloudevents client instrumented with OpenTelemetry
	ceClient, err := otelObs.NewClientHTTP([]cehttp.Option{}, []client.Option{})
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	ctx := cloudevents.ContextWithTarget(context.Background(), "http://localhost:8082/event")
	ctx = cloudevents.WithEncodingStructured(ctx)

	if result := ceClient.Send(ctx, event); cloudevents.IsUndelivered(result) {
		fmt.Fprintf(w, "Failed to send event to intermediary")
	}
}
