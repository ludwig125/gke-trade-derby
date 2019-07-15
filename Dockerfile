FROM golang:1.8-alpine
ADD . /go/src/trade-derby
RUN go install trade-derby

FROM alpine:latest
COPY --from=0 /go/bin/trade-derby .
ENV PORT 8080
CMD ["./trade-derby"]
