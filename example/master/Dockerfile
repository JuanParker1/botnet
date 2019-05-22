FROM alpine:latest AS master
RUN apk add --update bash curl && rm -rf /var/cache/apk/*
COPY master_linux /bin/master_linux
EXPOSE 443
EXPOSE 80
CMD ["/bin/master_linux"]
