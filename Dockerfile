FROM golang as builder

WORKDIR /hooker

COPY . .

RUN go get gopkg.in/go-playground/webhooks.v5/github
RUN go build .

FROM alpine

COPY --from=builder /hooker/hooker /bin/hooker

ENTRYPOINT /bin/hooker
