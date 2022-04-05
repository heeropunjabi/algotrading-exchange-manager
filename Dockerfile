##
## Build
##
FROM golang:1.18-alpine AS build

WORKDIR /
COPY go.mod ./
COPY go.sum ./
COPY main.go ./
RUN go mod download

COPY app/ ./app 
RUN rm -f ./app/zfiles/log/*.log

RUN go build -o /algoexmgr


##
## Deploy
##
FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=build /algoexmgr /algoexmgr

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/algoexmgr"]

# CMD [ "/algoexmgr" ]
