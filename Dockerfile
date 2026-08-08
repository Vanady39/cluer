# syntax=docker/dockerfile:1

# Which binary to build. Both services live in one module and share the domain
# packages, so one Dockerfile with an argument beats two that drift apart.
ARG SERVICE=onboarding

# ---------------------------------------------------------------- build
FROM golang:1.26-alpine AS build

WORKDIR /src

# Dependencies are copied on their own so the module download layer is reused
# whenever only source files change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE
# CGO off gives a static binary, which is what makes the scratch-like runtime
# stage below possible.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /out/service ./cmd/${SERVICE}

# ---------------------------------------------------------------- runtime
FROM alpine:3.22

# wget is used by the compose healthcheck; ca-certificates keeps outbound TLS
# working if the service ever has to call anything.
RUN apk add --no-cache ca-certificates wget \
    && adduser -D -u 10001 cluer

COPY --from=build /out/service /usr/local/bin/service

# Runs unprivileged: nothing here needs root, and a container that does not need
# it should not have it.
USER cluer

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/service"]
