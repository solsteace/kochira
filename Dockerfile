FROM golang:1.24.5-alpine3.21
WORKDIR /app

# Add `air` for HMR. Handy for dev environment
RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

CMD ["air"]
