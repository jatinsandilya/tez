############################
# STEP 1 build executable binary
# We specify the base image we need for our
# go application
############################
FROM golang:1.12.0-alpine3.9 AS builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

# We create an /app directory within our
# image that will hold our application source
# files

RUN mkdir /app

# We specify that we now wish to execute 
# any further commands inside our /app
# directory
WORKDIR /app

# Copy go mod & sum to download only updated dependencies
COPY go.mod .
COPY go.sum .


# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download

# We copy everything in the root directory
# into our /app directory
COPY . .

# we run go build to compile the binary
# executable of our Go program
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /main .

# List out directory
RUN ls -l

############################
# STEP 2 To build a small image
############################
FROM alpine:latest

# Copy our static executable.
COPY --from=builder /main ./

RUN chmod +x ./main

# Run our 
# newly created binary executable
ENTRYPOINT [ "./main" ]

EXPOSE 3000
