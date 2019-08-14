FROM golang:1.12.4-alpine3.9 as build-step

# for go mod download
RUN apk add --update --no-cache ca-certificates git

RUN mkdir /go-app
WORKDIR /go-app
COPY go.mod .
COPY go.sum .

RUN go mod download
COPY . .

RUN CGO_ENABLED=0 go build -o /go/src/trade-derby

FROM alpine:latest
COPY --from=build-step /go/src/trade-derby /go/src/trade-derby
ENV PORT 8080
CMD ["./trade-derby"]
