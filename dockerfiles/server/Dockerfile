FROM golang:1.17 as build

WORKDIR /build

COPY . .

RUN go build -o /app ./cmd/server

ENTRYPOINT ["/app"]