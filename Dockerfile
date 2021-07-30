ARG GOARCH=amd64

############################
FROM golang:1.15 AS builder
ARG GOARCH

WORKDIR /go/src/app

COPY ["go.mod", "./"]
RUN go mod download

# copy other sources & build
COPY ["./main.go", "/go/src/app/"]
ENV GOOS=linux
ENV GOARCH=${GOARCH}

RUN GOOS=linux CGO_ENABLED=0 go build -o /go/bin/upload-server


############################
FROM alpine:latest AS app
ARG GOARCH
ENV GOARCH=${GOARCH}

COPY --from=builder /go/bin/upload-server /usr/local/bin/

CMD ["/usr/local/bin/upload-server"]

VOLUME ["/storage/"]
ENV MAX_UPLOAD_SIZE=1073741824

EXPOSE 4500/TCP
