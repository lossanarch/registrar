FROM golang:1.12.13-alpine3.10 as gobuilder

WORKDIR /code

RUN apk --no-cache add ca-certificates \
    && update-ca-certificates

COPY . .

# CGO_ENABLED=0 stops C library linking, allowing running in scratch
# -installsuffix are not required
# the ldflags strip unnecessary and possibly security sensitive information
RUN set -ex && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -a -ldflags="-w -s"

FROM scratch

COPY --from=gobuilder /code/registrar /bin/
COPY --from=gobuilder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["registrar"]
