# Global ARG (set before the first build-stage), can be used in each build-stage.
ARG MAXMIND_LICENSE_KEY

FROM golang:1.13 AS BUILD
ARG MAXMIND_LICENSE_KEY

WORKDIR /opt

# create folder to store databases
RUN mkdir /opt/db

# download db if arg is not empty
COPY scripts/download-geolite2-city.sh .
RUN /opt/download-geolite2-city.sh $MAXMIND_LICENSE_KEY

#city state csv for Brazil
RUN curl https://raw.githubusercontent.com/chandez/Estados-Cidades-IBGE/master/Municipios.sql --output Municipios.sql
RUN awk -F ',' '{print "BR," $4 "," $5}' Municipios.sql | sed -e "s/''/#/g"  | sed -e "s/'//g" | sed -e "s/)//g" | sed -e "s/;//g" | sed -e s/", "/,/g | sed -e "s/#/'/g" > db/city-state.csv

WORKDIR /ruller-sample

COPY go.mod .
COPY go.sum .
ENV GO111MODULE on
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -ldflags="-w -s" -o ruller-sample

#Package
FROM alpine:3.11.5
ARG MAXMIND_LICENSE_KEY

ENV LOG_LEVEL "info"
ENV LISTEN_PORT "3000"
ENV LISTEN_ADDRESS "0.0.0.0"
ENV CITY_STATE_DB "/opt/city-state.csv"
ENV GEOLITE2_DB ${MAXMIND_LICENSE_KEY:+"/opt/Geolite2-City.mmdb"}

COPY --from=BUILD /ruller-sample /bin/
COPY --from=BUILD /opt/db/ /opt/

ADD scripts/startup.sh /

EXPOSE $LISTEN_PORT

CMD [ "/startup.sh" ]
