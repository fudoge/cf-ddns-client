FROM golang:1.25.12-bookworm AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/cfddns ./cmd/cfddns

FROM alpine:3.23.5

RUN adduser -D -H -u 10001 cfddns \
    && apk add --no-cache ca-certificates
COPY --from=builder /out/cfddns /usr/local/bin/cfddns

USER cfddns
ENTRYPOINT ["/usr/local/bin/cfddns"]
