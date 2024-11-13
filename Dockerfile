FROM golang:1-alpine AS build
WORKDIR /app

COPY go.mod go.[sum] ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go install -ldflags '-s -w -extldflags "-static"' github.com/go-delve/delve/cmd/dlv@latest
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main ./cmd/web

FROM alpine:3 AS prod
WORKDIR /app

COPY --from=build /go/bin/dlv /app/dlv
COPY --from=build /app/main /app/main

CMD ["/app/main"]
