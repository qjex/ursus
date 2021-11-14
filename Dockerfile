FROM golang:1.16.10-alpine3.14

RUN apk --no-cache add gcc libc-dev libpcap-dev iptables

ADD uwalker /build/uwalker
WORKDIR /build/uwalker

RUN go build -o uwalker

RUN cp /build/uwalker/uwalker /srv/uwalker
WORKDIR /srv
RUN mkdir data

COPY docker-init.sh /srv/init.sh
RUN chmod +x /srv/init.sh

CMD ["/srv/uwalker", "--sqlite", "data/db.sqlite", "-f", "data/subnets"]