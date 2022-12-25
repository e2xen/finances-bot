# Finances bot

It is a rather basic Telegram bot for finances management, implementing features like:
- handling of a new expense
- report generation of previously added expenses
- limiting your expenses
- all of that can be done in your preferred currency (currency conversion is done with an external API)

The app has 2 entrypoints, meant to be run as different instances:
- `cmd/bot/main.go`: the main bot functionality
- `cmd/reporter/main.go`: report generation as a separate program

Bot and reporter communicate as follows:
- user asks bot for a report
- bot sends a **Protobuf** message requesting the report from reporter through a **Kafka** topic
- reporter takes necessary data from a **PostgreSQL** database and generates the report
- then it sends the report to bot through **gRPC**
- bot sends the report to user

Other than that, reports are cached in **Memcached** to prevent regenerations.

## Tracing and Metrics

The app can send traces to **Jaeger** and implements `/metrics` route to facilitate **Prometheus** metrics collection.

## Docker

Although the app itself is not dockerized yet, all the necessary components can be found at the [docker compose](./docker-compose.yml) file.

## Testing

Basic testing of message handlers is done with the `testing` framework. Mocks are generated using [`minimock`](http://github.com/gojuno/minimock).

## Migrations

Migrations are performed with the [`migrate`](https://github.com/golang-migrate/migrate) tool (see [migrations](./migrations) directory).
