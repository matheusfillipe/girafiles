FROM golang:1.22
WORKDIR /app
COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-gs-ping
ARG TESTING
RUN \
  if [ "$TESTING" = "true" ]; then \
    apt-get update && apt-get install -y sqlite3; \
  fi

ENV PORT=8080
EXPOSE 8080

CMD ["go", "run", "main.go"]
