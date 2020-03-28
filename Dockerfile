FROM golang:1.10 as Builder

ENV PROJECT /go/src/github.com/pprasanthi/job-queue
ADD . $PROJECT
WORKDIR $PROJECT

RUN go get -u github.com/golang/dep/cmd/dep \
  && make clean \
  && make vendor \
  && make test \
  && go install ./...

FROM golang:1.10

COPY --from=Builder /go/bin/job-queue  /usr/local/bin/job-queue

ENTRYPOINT ["/usr/local/bin/job-queue"]


