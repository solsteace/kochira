FROM golang:1.24.5-alpine3.21
WORKDIR /app
COPY . .

# Add `air` for HMR. Handy for dev environment
RUN go install github.com/air-verse/air@latest

RUN go mod download

CMD ["air", "-c", "./.air.toml"]