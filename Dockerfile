FROM alpine
COPY /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY tetatetchatbot /
EXPOSE 8080
CMD ["/tetatetchatbot"]
