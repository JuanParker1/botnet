FROM alpine:latest AS slave
RUN apk add --update bash curl && rm -rf /var/cache/apk/*
COPY slave_app_linux /bin/slave_app_linux
CMD ["/bin/slave_app_linux"]
