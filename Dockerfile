FROM golang:1.22
WORKDIR /app
COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-gs-ping

ENV PORT=8080
EXPOSE 8080
CMD ["go", "run", "main.go"]
