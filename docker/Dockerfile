FROM golang:1.19

WORKDIR /go/src/kafka-consumer

COPY ./ ./

RUN go install ./pipeconsumer/cmd/kafka-pipe-consumer.go

CMD ["/go/bin/kafka-pipe-consumer"]