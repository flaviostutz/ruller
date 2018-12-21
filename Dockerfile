FROM golang:1.10 AS BUILD

RUN curl https://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz --output /opt/GeoLite2-City.tar.gz
RUN cd /opt && \
    tar -xvf GeoLite2-City.tar.gz && \
    mv GeoLite2-City_20181218/GeoLite2-City.mmdb /opt/Geolite2-City.mmdb && \
    rm -rf GeoLite2-City_20181218 && \
    rm GeoLite2-City.tar.gz

#doing dependency build separated from source build optimizes time for developer, but is not required
#install external dependencies first
ADD /main.go $GOPATH/src/ruller-sample/main.go
RUN go get -v ruller-sample

#now build source code
ADD ruller $GOPATH/src/ruller
RUN go get -v ruller

ADD ruller-sample $GOPATH/src/ruller-sample
RUN go get -v ruller-sample
#RUN go test -v ruller-sample


FROM golang:1.10

ENV LOG_LEVEL 'info'
ENV LISTEN_PORT '3000'
ENV LISTEN_ADDRESS '0.0.0.0'
ENV GEOLITE2_DB "/opt/Geolite2-City.mmdb"

COPY --from=BUILD /go/bin/* /bin/
COPY --from=BUILD /opt/Geolite2-City.mmdb /opt/
ADD startup.sh /

EXPOSE 3000

CMD [ "/startup.sh" ]
