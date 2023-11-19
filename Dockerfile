FROM golang:1.20.1 AS builder

COPY . /src
WORKDIR /src

RUN go env -w GOPROXY="https://goproxy.cn,direct"
RUN make build

FROM debian:stable-slim

COPY --from=builder /src/bin /app
COPY --from=builder /src/configs /data/conf
COPY --from=builder /src/polaris.yaml /app/polaris.yaml

WORKDIR /app

EXPOSE 8000
EXPOSE 9000
VOLUME /data/conf

CMD ["./user", "-conf", "/data/conf"]
