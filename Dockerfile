FROM golang:1.15 as go-builder
WORKDIR /cashflow
COPY go.mod go.sum /cashflow/
RUN go mod download
ADD . /cashflow/
RUN CGO_ENABLED=0 go build -o cashflow-server ./cmd/server

FROM alpine:3.11
LABEL org.opencontainers.image.source https://github.com/thcyron/cashflow
WORKDIR /cashflow
RUN apk add --no-cache ca-certificates
COPY --from=go-builder /cashflow/cashflow-server /usr/bin/cashflow-server
COPY testdata testdata
ENTRYPOINT ["cashflow-server"]
