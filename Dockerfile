# Stage 1: Build executable
FROM golang:1.12-alpine as builder

LABEL maintainer="Karthikeyan Govindaraj <github.gkarthiks@gmail.com>"

WORKDIR /go/src/github.com/gkarthiks/container-resource-exporter
COPY container_resource_exporter.go .

RUN apk update && apk add git ca-certificates && rm -rf /var/cache/apk/*

#RUN go get -v -t  .
RUN set -x && \
    go get github.com/sirupsen/logrus && \  
    go get github.com/prometheus/client_golang/prometheus && \
    go get k8s.io/api/core/v1 && \
    go get k8s.io/apimachinery/pkg/apis/meta/v1 && \
    go get k8s.io/client-go/kubernetes && \
    go get k8s.io/metrics/pkg/apis/metrics/v1beta1 && \
    go get k8s.io/metrics/pkg/client/clientset/versioned

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o /cr-exporter

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /cr-exporter .
EXPOSE 9000
ENTRYPOINT [ "/cr-exporter" ]