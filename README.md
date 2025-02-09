# OpenTelemetry Dice Service

A [dice](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/examples/dice) rolling service demonstrating OpenTelemetry instrumentation with Grafana Alloy and Grafana Cloud integration. 
The service provides HTTP endpoints for rolling dice and collects telemetry data including logs.

## Features

- RESTful endpoints for dice rolling
- Complete OpenTelemetry instrumentation:
  - Logs using Agoda's OpenTelemetry logs implementation
  - Metrics tracking roll counts and values
  - Distributed tracing for request flows
- Kubernetes deployment with Kind
- Grafana Cloud integration for observability

## Prerequisites

- Go 1.23+
- Docker
- Kind (Kubernetes in Docker)
- kubectl
- Helm
- make
- Grafana Cloud account and credentials

## Installation
Provide following environment variables in makefile:
- GRAFANA_ENDPOINT
- GRAFANA_USERNAME
- GRAFANA_PASSWORD

## Usage

```bash
make <target>
```

## Cleanup

```bash
make clean
```


