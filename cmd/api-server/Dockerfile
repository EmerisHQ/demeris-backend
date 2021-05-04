FROM golang:1.16 as builder

ARG GIT_TOKEN

RUN go env -w GOPRIVATE=github.com/allinbits/*
RUN git config --global url."https://git:${GIT_TOKEN}@github.com".insteadOf "https://github.com"

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOPROXY=direct make api-server

FROM alpine:latest

RUN apk --no-cache add ca-certificates mailcap && addgroup -S app && adduser -S app -G app
USER app
WORKDIR /app
COPY --from=builder /app/api-server /usr/local/bin/api-server
ENTRYPOINT ["/usr/local/bin/api-server"]
