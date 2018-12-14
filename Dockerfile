FROM golang:1.10 AS BUILD

#doing dependency build separated from source build optimizes time for developer, but is not required
#install external dependencies first
ADD /ruller-sample.go $GOPATH/src/ruller-sample/ruller-sample.go
RUN go get -v ruller-sample

#now build source code
ADD ruller-sample $GOPATH/src/ruller-sample
RUN go get -v ruller-sample
#RUN go test -v ruller-sample


FROM golang:1.10

ENV LOG_LEVEL 'info'
ENV LISTEN_PORT '3000'
ENV LISTEN_ADDRESS '0.0.0.0'

COPY --from=BUILD /go/bin/* /bin/
ADD startup.sh /

EXPOSE 3000

CMD [ "/startup.sh" ]
