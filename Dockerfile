FROM golang:alpine AS builder

WORKDIR /app

COPY . .

RUN apk add --no-cache build-base

RUN go mod download

RUN CGO_ENABLED=1 go build -a -o /app/build/bangs ./main.go

FROM alpine:latest

WORKDIR /usr/src/bangs

COPY --from=builder /app/build/bangs /usr/src/crons/bangs

RUN chmod +x /usr/src/bangs/bangs

ARG ENV_VARIABLE
ENV ENV_VARIABLE=${ENV_VARIABLE}

EXPOSE 8080

ENTRYPOINT ["/usr/src/bangs/bangs"]

CMD []
