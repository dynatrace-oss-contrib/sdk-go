package main

import (
	"context"
	"log"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	otelObs "github.com/cloudevents/sdk-go/observability/opentelemetry/v2/client"
	cenats "github.com/cloudevents/sdk-go/protocol/nats/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/cloudevents/sdk-go/v2/protocol"

	"github.com/cloudevents/sdk-go/samples/http/otel-sender-intermidiary-nats/instrumentation"
)

var tracer trace.Tracer

const (
	serviceName = "consumer"
)

func main() {
	shutdown := instrumentation.InitOTelSdk(serviceName)
	tracer = otel.Tracer(serviceName + "-main")
	defer shutdown()

	ctx := context.Background()

	// create the cloudevents client instrumented with OpenTelemetry
	p, err := cenats.NewConsumer("http://localhost:4222", "example.type", cenats.NatsOptions())
	if err != nil {
		log.Printf("Failed to create nats protocol, %v", err)
	}
	defer p.Close(ctx)

	c, err := cloudevents.NewClient(p, client.WithObservabilityService(otelObs.NewOTelObservabilityService()))
	if err != nil {
		log.Fatalf("failed to create client, %s", err.Error())
	}

	log.Println("Consumer started listening...")
	for {
		if err := c.StartReceiver(ctx, handleReceivedEvent); err != nil {
			log.Printf("failed to start nats receiver, %s", err.Error())
		}
	}
}

func handleReceivedEvent(ctx context.Context, event cloudevents.Event) protocol.Result {
	// fire a http request to make sure the propagation is working -
	// this span will be a child of the CloudEvents processing span
	ctx, childSpan := tracer.Start(ctx, "externalHttpCall", trace.WithAttributes(attribute.String("id", "123")))
	defer childSpan.End()

	// manually creating a http client instrumented with OpenTelemetry to make an external request
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	req, _ := http.NewRequestWithContext(ctx, "GET", "https://cloudevents.io/", nil)

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	_ = res.Body.Close()

	return nil
}
