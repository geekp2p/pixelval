# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS build
ENV GOTOOLCHAIN=auto
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /pixelval ./cmd/pixelval

FROM alpine:3.20
WORKDIR /app
RUN adduser -D pixelval
USER pixelval
COPY --from=build /pixelval ./pixelval
EXPOSE 8081 4001 4003
ENTRYPOINT ["./pixelval"]