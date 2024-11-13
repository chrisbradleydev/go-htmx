FROM golang:1-alpine AS build
WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux go install -ldflags '-s -w -extldflags "-static"' github.com/go-delve/delve/cmd/dlv@latest

COPY go.mod go.[sum] ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main ./cmd/web

FROM alpine:3 AS dev
WORKDIR /app

COPY --from=build /app/main /app/main
COPY --from=build /go/bin/dlv /app/dlv

CMD ["/app/main"]

FROM alpine:3 AS prod
WORKDIR /app

COPY --from=build /app/main /app/main

CMD ["/app/main"]
