# source: docker.com 
FROM golang:latest

RUN apt -y update; apt install -y net-tools

# Set destination for COPY
WORKDIR /app

# # Download Go modules
# COPY go.mod ./
# RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/reference/dockerfile/#copy
COPY *.go ./