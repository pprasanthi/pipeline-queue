FROM golang:1.10 as Builder

ENV PROJECT /go/src/github.com/pprasanthi/pipeline-queue
ADD . $PROJECT
WORKDIR $PROJECT

RUN go get -u github.com/golang/dep/cmd/dep \
  && make clean \
  && make vendor \
  && make test \
  && go install ./...

FROM golang:1.10

COPY --from=Builder /go/bin/pipeline-queue  /usr/local/bin/pipeline-queue

ENTRYPOINT ["/usr/local/bin/pipeline-queue"]
