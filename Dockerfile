# syntax=docker/dockerfile:1

FROM golang:1.17 AS build

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /usr/local/bin/file_exporter .

### Deploy
FROM alpine:latest

WORKDIR /

COPY --from=build /usr/local/bin/file_exporter /usr/local/bin/file_exporter

EXPOSE 9393
ENTRYPOINT [ "/usr/local/bin/file_exporter" ]