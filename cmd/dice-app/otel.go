// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs/otlplogsgrpc"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs/otlplogshttp"
	"github.com/agoda-com/opentelemetry-logs-go/logs"
	sdklog "github.com/agoda-com/opentelemetry-logs-go/sdk/logs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// Package-level logger
var logger logs.Logger

func InitProvider(serverName string) func() {
	ctx := context.Background()
	logres := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serverName),
		attribute.String("team", os.Getenv("TEAM")),
		attribute.String("prio", os.Getenv("PRIO")),
	)
	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serverName),
			attribute.String("environment", os.Getenv("go_env")),
			attribute.String("team", os.Getenv("TEAM")),
			attribute.String("prio", os.Getenv("PRIO")),
		),
	)
	HandleErr(err, "failed to create resource")

	// Get both gRPC and HTTP endpoints
	grpcEndpoint, ok := os.LookupEnv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if !ok {
		grpcEndpoint = "0.0.0.0:4317"
	}
	httpEndpoint, ok := os.LookupEnv("OTEL_EXPORTER_OTLP_HTTP_ENDPOINT")
	if !ok {
		httpEndpoint = "0.0.0.0:4318"
	}

	// Create gRPC exporter
	grpcExp, err := otlplogs.NewExporter(ctx,
		otlplogs.WithClient(
			otlplogsgrpc.NewClient(
				otlplogsgrpc.WithEndpoint(grpcEndpoint),
				otlplogsgrpc.WithInsecure(),
			),
		),
	)
	HandleErr(err, "Failed to create gRPC exporter")

	// Create HTTP exporter
	httpExp, err := otlplogs.NewExporter(ctx,
		otlplogs.WithClient(
			otlplogshttp.NewClient(
				otlplogshttp.WithEndpoint(httpEndpoint),
				otlplogshttp.WithInsecure(),
			),
		),
	)
	HandleErr(err, "Failed to create HTTP exporter")

	// Use both exporters
	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithBatcher(grpcExp),
		sdklog.WithBatcher(httpExp),
		sdklog.WithResource(logres),
	)
	logger = loggerProvider.Logger("demo-logger")
	now := time.Now()
	sev := logs.SeverityNumber(9)
	logger.Emit(logs.NewLogRecord(logs.LogRecordConfig{
		Timestamp:      &now,
		BodyAny:        "Initialized OTEL Provider from otel.go",
		SeverityNumber: &sev,
	}))
	traceClient := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(grpcEndpoint),
	)
	traceExp, err := otlptrace.New(ctx, traceClient)
	HandleErr(err, "Failed to create the collector trace exporter")
	bsp := sdktrace.NewBatchSpanProcessor(traceExp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(GetSampler()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetTracerProvider(tracerProvider)
	return func() {
		cxt, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err := traceExp.Shutdown(cxt); err != nil {
			otel.Handle(err)
		}
	}
}
func HandleErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}

// Helper function to define sampling.
// When in development mode, AlwaysSample is defined,
// otherwise, sample based on Parent and IDRatio will be used.
func GetSampler() sdktrace.Sampler {
	ENV := os.Getenv("GO_ENV")
	switch ENV {
	case "development":
		return sdktrace.AlwaysSample()
	case "production":
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.5))
	default:
		return sdktrace.AlwaysSample()
	}
}
func Log(message string, severity logs.SeverityNumber) {
	now := time.Now()
	logger.Emit(logs.NewLogRecord(logs.LogRecordConfig{
		Timestamp:      &now,
		BodyAny:        message,
		SeverityNumber: &severity,
	}))
}
func GetLogger() logs.Logger {
	return logger
}
