FROM golang:1.12 as build-env
WORKDIR /app
ADD . /app
RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo ./cmd/backend-challenge

FROM gcr.io/distroless/base
ADD ./web /app/web
COPY --from=build-env /app/backend-challenge /app
WORKDIR /app
EXPOSE 8000
CMD ["/app/backend-challenge"]