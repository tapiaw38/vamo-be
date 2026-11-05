FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o /out/vamo-backend ./cmd/api

FROM alpine:3.22
WORKDIR /app
COPY --from=build /out/vamo-backend /app/vamo-backend
COPY docs /app/docs
EXPOSE 8080
CMD ["/app/vamo-backend"]
