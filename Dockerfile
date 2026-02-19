# syntax=docker/dockerfile:1.7

FROM golang:1.22 AS builder
WORKDIR /src

COPY go.mod ./
# go.sum may not exist in fresh scaffolds; copy conditionally by wildcard
COPY go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /bin/product-catalog ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /bin/product-catalog /app/product-catalog

ENV GRPC_ADDR=:50051
EXPOSE 50051
USER nonroot:nonroot
ENTRYPOINT ["/app/product-catalog"]
