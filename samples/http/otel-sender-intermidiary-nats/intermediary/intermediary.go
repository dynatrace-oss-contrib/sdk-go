package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/cloudevents/sdk-go/samples/http/otel-sender-intermidiary-nats/instrumentation"
)

const (
	serviceName = "intermediary"
)

func main() {
	shutdown := instrumentation.InitOTelSdk(serviceName)
	defer shutdown()

	otelHandler := otelhttp.NewHandler(http.HandlerFunc(handleReceivedEvent), "Received CloudEvent")
	http.Handle("/event", otelHandler)

	log.Println("Intermediary listening on localhost:8081")
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func handleReceivedEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	event := cloudevents.NewEvent()

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&event)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "error ocurred"}`))
	}

	// forward to NATS
	p, err := cenats.NewSender("http://localhost:4222", event.Context.GetType(), cenats.NatsOptions())
	if err != nil {
		log.Printf("Failed to create nats protocol, %v", err)
	}

	c, err := cloudevents.NewClient(p)
	if err != nil {
		log.Printf("Failed to create client, %v\n", err)
	}

	ctx := cloudevents.WithEncodingStructured(context.Background())
	c.Send(ctx, event)

	w.WriteHeader(http.StatusOK)
}
