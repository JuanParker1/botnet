FROM alpine:latest AS ccserver
RUN apk add --update bash curl && rm -rf /var/cache/apk/*
COPY ccserver_linux /bin/ccserver_linux
EXPOSE 443
EXPOSE 80
CMD ["/bin/ccserver_linux"]
