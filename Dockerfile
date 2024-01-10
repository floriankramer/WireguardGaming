FROM golang:1.21-alpine3.19 as builder

WORKDIR /opt/app
COPY . .
RUN go build

FROM alpine:3.19

RUN apk add --no-cache iptables iproute2 wireguard-tools

COPY --from=builder /opt/app/wireguard_gaming /usr/local/bin/wireguard_gaming

ENTRYPOINT wireguard_gaming
