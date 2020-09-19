# Global ARG (set before the first build-stage), can be used in each build-stage.
ARG MAXMIND_LICENSE_KEY

FROM golang:1.14.3-alpine3.11 AS BUILD

RUN apk add curl

WORKDIR /opt

# download db if arg is not empty
ADD /sample/download-geolite2-city.sh /opt/
RUN /opt/download-geolite2-city.sh $MAXMIND_LICENSE_KEY

#city state csv for Brazil
RUN curl https://raw.githubusercontent.com/chandez/Estados-Cidades-IBGE/master/Municipios.sql --output Municipios.sql
RUN awk -F ',' '{print "BR," $4 "," $5}' Municipios.sql | sed -e "s/''/#/g"  | sed -e "s/'//g" | sed -e "s/)//g" | sed -e "s/;//g" | sed -e s/", "/,/g | sed -e "s/#/'/g" > /opt/city-state.csv

WORKDIR /ruller/sample

ADD /go.mod /ruller/
ADD /go.sum /ruller/
ADD /sample/go.mod /ruller/sample/
ADD /sample/go.sum /ruller/sample/

# ENV GO111MODULE on
RUN go mod download

COPY / /ruller/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -ldflags="-w -s" -o /ruller-sample




FROM alpine:3.11.5
ENV LOG_LEVEL "info"

COPY --from=BUILD /ruller-sample /bin/
COPY --from=BUILD /opt/ /opt/

ADD /sample/startup.sh /

EXPOSE 3000

CMD [ "/startup.sh" ]
