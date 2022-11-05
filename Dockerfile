FROM golang:1.19

WORKDIR /go/src/kafka-consumer

COPY ./ ./

RUN go install 

CMD ["/go/bin/kafka-consumer"]