FROM alpine:latest AS bot
RUN apk add --update bash curl && rm -rf /var/cache/apk/*
COPY bot_linux /bin/bot_linux
CMD ["/bin/bot_linux"]
