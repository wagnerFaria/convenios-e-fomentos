FROM golang:1.27 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/cli ./cmd/cli

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /out/cli /
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/api"]
