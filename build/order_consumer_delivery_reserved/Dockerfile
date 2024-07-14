FROM golang:1.21 as build

ENV BIN_FILE /opt/app/bin
ENV CODE_DIR /go/src/

WORKDIR ${CODE_DIR}

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY .. ${CODE_DIR}

ARG LDFLAGS
RUN CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o ${BIN_FILE} cmd/order_consumer_delivery_reserved/*

FROM alpine:latest

LABEL SERVICE="order_consumer_delivery_reserved"
LABEL MAINTAINERS="zaytcev.com"

ENV BIN_FILE /opt/app/bin
COPY --from=build ${BIN_FILE} ${BIN_FILE}

ENV CONFIG_FILE /etc/app/config.json
COPY ./configs/order_consumer_delivery_reserved/config.json ${CONFIG_FILE}

CMD ${BIN_FILE} -config ${CONFIG_FILE}
