FROM golang:1.7

COPY main.go /go/src/simpleBlockChain/

WORKDIR /go/src/simpleBlockChain

RUN ["go", "build", "-o", "simplechain", "main.go"]

ENV HTTP_PORT=9090 P2P_PORT=7676

EXPOSE $HTTP_PORT $P2P_PORT

ENTRYPOINT ["./simplechain"]
