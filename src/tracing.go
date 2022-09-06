package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc/credentials"

	"go.opentelemetry.io/otel"
	otlp "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	otlpgrpc "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func InitializeTracing(ctx context.Context) *otlp.Exporter {

	// stdout exporter
	// uncomment below (four lines here, and one on line ~53) to see your events printed to the console
	// std, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	// if err != nil {
	//   log.Fatal(err)
	// }

	serviceName, _ := os.LookupEnv("SERVICE_NAME")
	os.Stderr.WriteString(fmt.Sprintf("Sending as service name %s\n", serviceName))

	hny := connectToHoneycomb(ctx)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String(serviceName))),
		// uncomment (one line here, plus four above) to see your events printed to the console
		// sdktrace.WithSyncer(std),
		sdktrace.WithBatcher(hny))

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return hny
}

func connectToHoneycomb(ctx context.Context) *otlp.Exporter {
	apikey, _ := os.LookupEnv("HONEYCOMB_API_KEY")
	os.Stderr.WriteString(fmt.Sprintf("Sending to Honeycomb with API Key <%s>\n", apikey))

	// set up grpc
	driver := otlpgrpc.NewClient(
		otlpgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
		otlpgrpc.WithEndpoint("api.honeycomb.io:443"),
		otlpgrpc.WithHeaders(map[string]string{
			"x-honeycomb-team": apikey,
		}),
	)
	hny, err := otlp.New(ctx, driver)
	if err != nil {
		log.Fatal(err)
	}

	return hny
}
