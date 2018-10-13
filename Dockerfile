FROM golang:latest as build

RUN go get -u golang.org/x/vgo

WORKDIR $GOPATH/src/github.com/ushmodin/tetatetchatbot
COPY . .
RUN CGO_ENABLED=0 GO111MODULE=on vgo build && cp ./tetatetchatbot /

FROM alpine
COPY --from=build /tetatetchatbot /
EXPOSE 8080
CMD ["/tetatetchatbot"]
