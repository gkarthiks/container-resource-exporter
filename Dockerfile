# Stage 1: Build executable
FROM golang:1.14-alpine as builder

LABEL maintainer="Karthikeyan Govindaraj <github.gkarthiks@gmail.com>"

WORKDIR /go/src/github.com/gkarthiks/container-resource-exporter
COPY container_resource_exporter.go .

COPY go.mod .
COPY go.sum .

RUN apk update && apk add git ca-certificates && rm -rf /var/cache/apk/*

RUN go get -v -t  .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o /cr-exporter

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /cr-exporter .
EXPOSE 9000
ENTRYPOINT [ "/cr-exporter" ]
