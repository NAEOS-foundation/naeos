# Microservices Go Starter

Go microservices starter project with REST API and event-driven communication,
generated from the `microservices-go` NAEOS template.

## Layout

```
cmd/api/               REST API entry point
internal/order/        Order domain logic
internal/config/       Env-based configuration
proto/order/v1/        gRPC contract (stub, ready for protoc)
spec.yaml              NAEOS source spec for this service
naeos.yaml             NAEOS pipeline configuration
```

## Quick start

```bash
# Generate code from the NAEOS spec
naeos generate spec.yaml --config naeos.yaml

# Run the API
go run ./cmd/api
curl -X POST localhost:8080/orders -d '{"id":"1","total":100}'
```

## NAEOS pipeline

`naeos.yaml` wires the spec through the full NAEOS pipeline
(parse → normalize → resolve → build → validate → schedule → generate →
render). Run it with:

```bash
naeos run spec.yaml
```

## Event-driven services

The spec declares `order-worker` (kind: worker). Wire it to NATS and subscribe
to the `OrderCreated` topic to drive payment and inventory modules.

## Publishing as a template

This directory is a valid template: it has a `template.yaml` manifest and a
README. Publish it to the NAEOS template registry:

```bash
naeos template publish .
```
