FROM golang:1.18 AS builder
WORKDIR /build
COPY . .
RUN go mod tidy && \
    go build .

FROM ubuntu:20.04 AS eth_balances_counter
WORKDIR /bin
RUN apt update && apt install -y ca-certificates && update-ca-certificates && \
    openssl dhparam -out /etc/ssl/certs/dhparam.pem 2048
COPY --from=builder /build/eth_balances_counter .
ENTRYPOINT ["/bin/eth_balances_counter"]