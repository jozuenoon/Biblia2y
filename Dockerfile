FROM scratch
COPY bin/server .
COPY config/config.toml /config/
ADD data data
COPY ./db-empty-dir /db

ADD https://github.com/golang/go/raw/master/lib/time/zoneinfo.zip /zoneinfo.zip
ENV ZONEINFO /zoneinfo.zip

ADD https://curl.haxx.se/ca/cacert.pem /etc/ssl/certs/ca-certificates.crt

CMD ["/server"]