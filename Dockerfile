FROM golang:1.10 AS BUILD

#doing dependency build separated from source build optimizes time for developer, but is not required
#install external dependencies first
ADD /ruller.go $GOPATH/src/ruller/ruller.go
RUN go get -v ruller

#now build source code
ADD ruller $GOPATH/src/ruller
RUN go get -v ruller
#RUN go test -v ruller-sample


FROM golang:1.10

ENV LOG_LEVEL 'info'
ENV LISTEN_PORT '3000'
ENV LISTEN_ADDRESS '0.0.0.0'

COPY --from=BUILD /go/bin/* /bin/
ADD startup.sh /

EXPOSE 3000

CMD [ "/startup.sh" ]
