# Build the manager binary
FROM golang:1.10.3 as builder

# Copy in the go src
WORKDIR /go/src/github.com/cvgw/rds-aurora-operator
COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor/ vendor/
COPY hack/ hack/

# Build
RUN go generate ./pkg/...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager github.com/cvgw/rds-aurora-operator/cmd/manager

# Copy the controller-manager into a thin image
FROM ubuntu:latest
WORKDIR /root/
COPY --from=builder /go/src/github.com/cvgw/rds-aurora-operator/manager .
ENTRYPOINT ["./manager"]
