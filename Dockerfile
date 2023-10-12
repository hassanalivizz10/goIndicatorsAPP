# First stage - builder
FROM golang:1.16 as indicatorsApp

# Set the working directory
WORKDIR /app

# Copy the entire project into the container
COPY . .

# Enable Go modules
ENV GO111MODULE=on

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o indicatorsBuildAPP

# Second stage
FROM alpine:latest

# Set the working directory in the final image
WORKDIR /root/

# Install time zone data (if needed)
RUN apk add --no-cache tzdata

# Copy the built Go binary from the builder stage
COPY --from=indicatorsApp /app/indicatorsBuildAPP .

# Specify the command to run your Go application
CMD ["./indicatorsBuildAPP"]