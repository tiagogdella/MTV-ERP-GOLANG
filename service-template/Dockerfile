# ----- build stage -----
FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/server

# ----- final stage -----
FROM gcr.io/distroless/static-debian12

COPY --from=builder /bin/server /bin/server

EXPOSE 8080

ENTRYPOINT ["/bin/server"]