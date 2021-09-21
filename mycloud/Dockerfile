FROM golang:1.17-buster AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download
RUN go mod tidy

COPY *.go ./

RUN go get
RUN go build -o /mycloud

##
## Deploy
##
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /mycloud /mycloud

EXPOSE 8081

USER nonroot:nonroot

ENTRYPOINT ["/mycloud"]