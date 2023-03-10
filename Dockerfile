FROM golang:1.20-alpine3.17 as builder

WORKDIR /src/

COPY go.mod go.sum /src/

RUN go mod download

COPY cmd /src/cmd
COPY pkg /src/pkg

RUN go build cmd/main.go

FROM alpine:3.17

COPY --from=builder /src/main /root/

WORKDIR /root/

ENTRYPOINT [ "/root/main" ]