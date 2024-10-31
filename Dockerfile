# syntax=docker/dockerfile:1
ARG GO_VERSION=latest
ARG GOLANGCI_LINT_VERSION=latest-alpine

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS deps
ENV GOMODCACHE=/go/pkg/mod/
ENV GOCACHE=/.cache/go-build/
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=${GOMODCACHE} \
    --mount=type=cache,target=${GOCACHE} \
    go mod download -x

FROM deps AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
COPY . .
RUN --mount=type=cache,target=${GOMODCACHE} \
    --mount=type=cache,target=${GOCACHE} \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build ./...

FROM scratch AS processor
WORKDIR /app
COPY --from=build /go/bin/receipt_processor .
EXPOSE 8080
ENTRYPOINT [ "./processor", "-addr=:8080" ]