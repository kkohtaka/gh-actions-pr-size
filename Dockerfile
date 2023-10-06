FROM golang:1.21.2 AS builder
COPY . /src/github.com/kkohtaka/gh-actions-pr-size
WORKDIR /src/github.com/kkohtaka/gh-actions-pr-size
RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on \
  go build \
  -a \
  -o /bin/pr-size \
  /src/github.com/kkohtaka/gh-actions-pr-size/cmd/gh-actions-pr-size/

FROM alpine:3.18.2 as certs-installer
RUN apk add --update ca-certificates

FROM scratch
COPY --from=builder /bin/pr-size /bin/pr-size
COPY --from=certs-installer /etc/ssl/certs /etc/ssl/certs
ENTRYPOINT ["/bin/pr-size"]
CMD [""]