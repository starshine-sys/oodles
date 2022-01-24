FROM golang:latest AS builder

WORKDIR /build
COPY . ./
RUN go mod download
ENV CGO_ENABLED 0
RUN go build -v -o oodles -ldflags="-X github.com/starshine-sys/oodles/common.Version=`git rev-parse --short HEAD`" .

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /build/oodles oodles

CMD ["/app/oodles"]
