FROM golang:1.7
ADD . /go/src/github.com/travisjeffery/burrow-stats
RUN go install github.com/travisjeffery/burrow-stats/cmd/burrow-stats
ENTRYPOINT ["/go/bin/burrow-stats"]
