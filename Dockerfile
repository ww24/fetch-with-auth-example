FROM golang:1.21 AS build

WORKDIR $GOPATH/src/github.com/ww24/fetch-with-auth-example
COPY . .

ENV CGO_ENABLED=0
RUN go build -buildmode pie -ldflags "-w" -o /usr/local/bin/server .

FROM gcr.io/distroless/base:nonroot

COPY --from=build /usr/local/bin/server /usr/local/bin/server
ENTRYPOINT [ "server" ]
