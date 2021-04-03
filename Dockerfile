############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
RUN apk update && apk add git
WORKDIR /myapp
ENV GO111MODULE=on
COPY go.mod go.sum ./
RUN go mod download
COPY . /myapp
RUN ls .
# Build the binary.
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GO111MODULE=on go build -ldflags="-w -s" -o /go/bin/myapp main.go

############################
# STEP 2 build a small image
############################
# FROM scratch
FROM alpine:latest  
RUN apk --no-cache add ca-certificates
# Import the user and group files from the builder.
COPY --from=builder /etc/passwd /etc/passwd

WORKDIR /root/
# Copy our static executable.
COPY --from=builder /go/bin/myapp .
COPY --from=builder /myapp/config.yml .
COPY --from=builder /myapp/service/user/config.yml ./service/user/config.yml
COPY --from=builder /myapp/service/user/client/config.yml ./service/user/client/config.yml

# gRPC server port
EXPOSE 9090
# proxy gateway
EXPOSE 8000

# CMD exec /bin/sh -c "trap : TERM INT; (while true; do sleep 1000; done) & wait"
ENTRYPOINT ["./myapp"]
CMD [ "grpc-gateway" ]