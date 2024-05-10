# Use the official Golang image to create a build artifact.
# This image is based on Debian and includes Golang version 1.19.
FROM golang:1.19 as builder

# Set the working directory outside $GOPATH to enable the go modules feature
WORKDIR /app

# Copy the go.mod and go.sum files to download the dependencies
# This is to leverage Docker cache to save re-downloading dependencies if unmodified
COPY go.mod go.sum ./

# Download the dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
ENV CGO_ENABLED=1
RUN go build -o networkHub ./cmd/main.go
RUN ./scripts/build_thor.sh

#### Use a Docker multi-stage build to create a lean production image.
#### Start with the Debian buster-slim image for a small footprint.
###FROM debian:buster-slim
##
### Set the working directory to /root/
##WORKDIR /root/
#
## Copy the built executable from the builder stage
#COPY --from=builder /app/networkHub .

# Set the necessary ports (assuming default HTTP API ports; adjust if different)
EXPOSE 80

# Command to run the executable
CMD ["./networkHub", "api"]
