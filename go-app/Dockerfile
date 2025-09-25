# Use Go official image as base
FROM golang:1.24.6-alpine


# Set working directory inside container
WORKDIR /app

# Copy Go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build the Go app
RUN go build -o main .

# Expose port your app listens on
EXPOSE 8082

# Command to run your app
CMD ["./main"]
