package main

import (
	"context"
	"fmt"
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

	hny, hnyConfigured := connectToHoneycomb(ctx)
	jaeger, jaegerConfigured := connectToJaeger(ctx)

	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String(serviceName))),
		// uncomment (one line here, plus four above) to see your events printed to the console
		// sdktrace.WithSyncer(std),
	}

	if hnyConfigured {
		opts = append(opts, sdktrace.WithBatcher(hny))
	}

	if jaegerConfigured {
		opts = append(opts, sdktrace.WithBatcher(jaeger))
	}

	tp := sdktrace.NewTracerProvider(
		opts...,
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return hny
}

func connectToJaeger(ctx context.Context) (*otlp.Exporter, bool) {
	loc, defined := os.LookupEnv("JAEGER_LOCATION")
	if !defined {
		os.Stderr.WriteString("To send traces to Jaeger, define JAEGER_LOCATION\n")
		return nil, false
	}

	os.Stderr.WriteString(fmt.Sprintf("Sending to Jaeger at <%s>\n", loc))

	// set up grpc
	driver := otlpgrpc.NewClient(
		otlpgrpc.WithInsecure(),
		otlpgrpc.WithEndpoint(fmt.Sprintf("%s:4317", loc)),
	)
	j, err := otlp.New(ctx, driver)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("or not, I guess %s", err))
		return nil, false
	}

	return j, true
}

func connectToHoneycomb(ctx context.Context) (*otlp.Exporter, bool) {
	apikey, defined := os.LookupEnv("HONEYCOMB_API_KEY")
	if !defined {
		os.Stderr.WriteString("To send traces to Honeycomb, define HONEYCOMB_API_KEY\n")
		return nil, false
	}

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
		os.Stderr.WriteString(fmt.Sprintf("or not, I guess %s", err))
		return nil, false
	}

	return hny, true
}
