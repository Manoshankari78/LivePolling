FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

FROM alpine:3.20
WORKDIR /app
RUN adduser -D -H appuser
COPY --from=build /out/server /app/server
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/server"]
