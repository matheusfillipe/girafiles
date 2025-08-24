FROM golang:1.25 as builder
ARG TESTING
RUN \
  if [ "$TESTING" = "true" ]; then \
    apt-get update && apt-get install -y sqlite3; \
  fi

WORKDIR /app
COPY . .
RUN go mod download

RUN CGO_ENABLED=1 GOOS=linux go build -o build/girafiles .

ENV PORT=8080
EXPOSE 8080

CMD ["./build/girafiles"]
