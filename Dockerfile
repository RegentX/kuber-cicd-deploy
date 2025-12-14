FROM golang:1.22 AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/app ./main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=build /out/app /app
USER nonroot:nonroot
EXPOSE 8080
ENV PORT=8080
ENV IMAGE_PATH=/data/image.png
ENTRYPOINT ["/app"]
