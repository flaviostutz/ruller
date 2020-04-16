FROM golang:1.10 AS BUILD

RUN curl https://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz --output /opt/GeoLite2-City.tar.gz
RUN cd /opt && \
    tar -xvf GeoLite2-City.tar.gz && \
    mv */GeoLite2-City.mmdb /opt/Geolite2-City.mmdb && \
    rm -rf GeoLite2-City_20181218 && \
    rm GeoLite2-City.tar.gz

#city state csv for Brazil
RUN curl https://raw.githubusercontent.com/chandez/Estados-Cidades-IBGE/master/Municipios.sql --output /opt/Municipios.sql
RUN awk -F ',' '{print "BR," $4 "," $5}' /opt/Municipios.sql | sed -e "s/''/#/g"  | sed -e "s/'//g" | sed -e "s/)//g" | sed -e "s/;//g" | sed -e s/", "/,/g | sed -e "s/#/'/g" > /opt/city-state.csv

#doing dependency build separated from source build optimizes time for developer, but is not required
#install external dependencies first
ADD /main.dep $GOPATH/src/ruller-sample/main.go
RUN go get -v ruller-sample

#now build source code
ADD ruller $GOPATH/src/ruller
RUN go get -v ruller

ADD sample $GOPATH/src/ruller-sample
RUN go get -v ruller-sample
#RUN go test -v ruller-sample


FROM golang:1.10

ENV LOG_LEVEL 'info'
ENV LISTEN_PORT '3000'
ENV LISTEN_ADDRESS '0.0.0.0'
ENV GEOLITE2_DB "/opt/Geolite2-City.mmdb"
ENV CITY_STATE_DB "/opt/city-state.csv"

COPY --from=BUILD /go/bin/* /bin/
COPY --from=BUILD /opt/Geolite2-City.mmdb /opt/
COPY --from=BUILD /opt/city-state.csv /opt/

ADD startup.sh /

EXPOSE 3000

CMD [ "/startup.sh" ]
