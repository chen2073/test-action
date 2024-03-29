FROM golang:1.22-alpine3.18 AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

FROM alpine:3.18 as production

COPY --from=builder /app/main /app/main

CMD ["/app/main"]