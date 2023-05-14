FROM golang:1.20-rc-alpine AS build
ENV GOPRIVATE="github.com/SekyrOrg/*"
ARG GITHUB_TOKEN

WORKDIR /src

COPY go.mod ./
COPY go.sum ./

# Set up Git and authenticate using the access token
RUN apk add --no-cache git && \
    git config --global url."https://${GITHUB_TOKEN}:@github.com/".insteadOf "https://github.com/"

RUN go mod download

COPY . .

RUN go build -o /src/out/forge cmd/main.go

FROM alpine:latest

COPY --from=build /src/out/forge /forge

