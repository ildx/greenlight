FROM golang:1.22 AS base
FROM base AS dev

RUN go install github.com/cosmtrek/air@latest

WORKDIR /app/
CMD ["air"]
