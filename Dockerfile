# syntax=docker/dockerfile:1
FROM golang:1.17-alpine
WORKDIR /app
# Download dependencied
COPY go.mod ./
COPY go.sum ./
RUN go mod download
# Copy files into /app and build project
COPY . /app
RUN go build -o /main
# Expose port
EXPOSE 8080
# Run main
CMD [ "/main" ]
