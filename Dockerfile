FROM golang:1.17 AS builder


RUN mkdir /app

ADD . /app

WORKDIR /app/

RUN CGO_ENABLED=0 GOOS=linux go build -o discover -trimpath -a -ldflags '-w -extldflags "-static"'



FROM alpine:latest

WORKDIR /
COPY --from=builder /app/discover  .

EXPOSE 8080

CMD [ "/discover", "api" ]
