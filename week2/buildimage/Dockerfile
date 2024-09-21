FROM public.ecr.aws/docker/library/golang:1.21 AS builder
#FROM golang:1.21 AS builder

WORKDIR /opt
#mkdir example && curl https://gist.githubusercontent.com/abhishekkr/3beebbc1db54b3b54914/raw/16f13fe1d4d222df0ff21db0e1265c5b583c7477/tcp_server.go > example/main.go
COPY ./example .
RUN CGO_ENABLED=0 go build -o /bin/example ./main.go

FROM scratch
COPY --from=builder /bin/example /bin/example
CMD ["/bin/example"]
