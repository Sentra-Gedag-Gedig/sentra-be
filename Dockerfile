FROM golang:1.21-alpine AS builder

WORKDIR /app

ENV GOTOOLCHAIN=auto

RUN apk --no-cache add git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/app

# Final stage with a smaller image and non-root user
FROM alpine:3.18

# Add necessary runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user and set home directory
RUN adduser -D -g '' -h /home/appuser appuser
WORKDIR /home/appuser

# Copy the binary and .env file with appropriate ownership
COPY --from=builder /app/main .
COPY --chown=appuser .env .

# Switch to the non-root user
USER appuser

# Expose the application port
EXPOSE 8080

# Command to run the executable
CMD ["./main"]