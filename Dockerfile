# Step 1: Build stage
FROM docker.arvancloud.ir/golang:1.25 AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp .

# Step 2: Final stage
FROM docker.arvancloud.ir/alpine:latest

# Set the working directory for the final image
WORKDIR /root/

# Copy the compiled binary from the builder image
COPY --from=builder /app/myapp .

# Expose the port your application will run on
EXPOSE 8080

# Command to run your application
CMD ["./myapp"]
