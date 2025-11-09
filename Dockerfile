FROM golang:latest AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o spi2c

RUN CGO_ENABLED=0 GOOS=linux go build -o spi2c

FROM alpine:latest

WORKDIR /root/

COPY --from=build /app/spi2c .

COPY /public ./public

EXPOSE 8081

CMD ["./spi2c"]
