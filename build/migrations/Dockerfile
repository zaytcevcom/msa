FROM golang:1.21 as build

ENV BIN_FILE /opt/app/bin
ENV CODE_DIR /go/src/

WORKDIR ${CODE_DIR}

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY .. ${CODE_DIR}

ARG LDFLAGS
RUN CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o ${BIN_FILE} cmd/migrations/*

FROM alpine:latest

LABEL SERVICE="demo-migrations"
LABEL MAINTAINERS="zaytcev.com"

ENV BIN_FILE /opt/app/bin
COPY --from=build ${BIN_FILE} ${BIN_FILE}

COPY migrations /migrations

CMD ${BIN_FILE} -dir migrations postgres "${DSN}" up