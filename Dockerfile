FROM alpine
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY tetatetchatbot /
EXPOSE 8080
CMD ["/tetatetchatbot"]
