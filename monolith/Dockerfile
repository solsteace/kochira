FROM golang:alpine AS base
WORKDIR /kochira
RUN mkdir ./monolith ./lib ./account ./link ./subscription
COPY ./monolith/go.* ./monolith
COPY ./account/go.* ./account
COPY ./link/go.* ./link
COPY ./subscription/go.* ./subscription
COPY ./lib/go.*  ./lib
RUN go work init && go work use ./monolith ./lib ./account ./link ./subscription


FROM alpine/curl:latest AS devtools
WORKDIR .
# goose: 
# - https://pressly.github.io/goose/installation/
# - https://github.com/pressly/goose
RUN curl -fsSL https://raw.githubusercontent.com/pressly/goose/master/install.sh | sh


FROM base AS devel
COPY --from=devtools /usr/local/bin /
RUN go install github.com/air-verse/air@latest
CMD ["air", "-c", "./monolith/.air.toml"]