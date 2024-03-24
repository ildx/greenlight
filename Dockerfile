FROM golang:1.22 AS base
FROM base AS dev

RUN go install github.com/cosmtrek/air@latest
RUN go install github.com/rakyll/hey@latest

WORKDIR /app/
CMD ["air"]
