# syntax=docker/dockerfile:1

FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod ./
COPY main.go index.html s.html app.html ./
COPY apps ./apps
COPY s ./s
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/i41eu-web .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/i41eu-web /i41eu-web
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/i41eu-web", "-listen", ":8080"]
