FROM golang:1.14 AS builder
COPY . /src/github.com/kkohtaka/gh-actions-pr-size
WORKDIR /src/github.com/kkohtaka/gh-actions-pr-size
RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on \
  go build \
  -a \
  -o /bin/pr-size \
  /src/github.com/kkohtaka/gh-actions-pr-size/cmd/gh-actions-pr-size/

FROM scratch
COPY --from=builder /bin/pr-size /bin/pr-size
ENTRYPOINT ["/bin/pr-size"]
CMD [""]