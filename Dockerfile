FROM golang:1.14 AS builder
ENV CGO_ENABLED 0
WORKDIR /go/src/app
ADD . .
RUN go build -o /auto-check-dups

FROM alpine:3.12
COPY --from=builder /auto-check-dups /auto-check-dups
CMD ["/auto-check-dups"]