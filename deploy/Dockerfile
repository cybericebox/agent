FROM golang:1.22.3-alpine AS builder
WORKDIR /build
RUN apk add gcc g++ --no-cache
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app -a -ldflags '-w -extldflags "-static"' /build/cmd/main.go

FROM alpine
WORKDIR /app

# copy the built binary into the image
COPY --from=builder /build/app /app/app

ENTRYPOINT ["/app/app"]
EXPOSE 5454