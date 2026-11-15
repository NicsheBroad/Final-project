FROM golang:latest AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /todo .

FROM ubuntu:latest
WORKDIR /app
COPY --from=build /todo /app/todo
COPY web /app/web
RUN mkdir -p /data

ENV TODO_PORT=7540
ENV TODO_DBFILE=/data/scheduler.db
ENV TODO_PASSWORD=12345

EXPOSE 7540
CMD ["/app/todo"]