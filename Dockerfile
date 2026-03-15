FROM golang:1.23 AS build
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /linkpulse ./cmd/server

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /linkpulse /linkpulse
COPY web ./web
EXPOSE 8080
ENTRYPOINT ["/linkpulse"]
