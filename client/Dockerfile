# source: docker.com 
FROM golang:latest

# Set destination for COPY
WORKDIR /app

# # Download Go modules
COPY go.mod ./
COPY go.sum ./
RUN go mod download;

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/reference/dockerfile/#copy
COPY *.go ./
RUN go build client.go rpc.go tree.go util.go