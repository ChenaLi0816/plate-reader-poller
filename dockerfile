FROM ubuntu:latest

WORKDIR /poller

COPY ./main /poller/

ENTRYPOINT ["./main"]
