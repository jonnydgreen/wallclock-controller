# Start by building the application.
FROM golang:1.13-buster as build

WORKDIR /go/src/app

# Get TPP first
COPY go.mod .
COPY go.sum .
RUN go mod download

# Run main build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/app

FROM gcr.io/distroless/base-debian10
COPY --from=build /go/bin/app /
CMD ["/app"]
