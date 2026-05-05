FROM golang:1.26 AS build
RUN useradd -u 10001 appuser
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /immich-smartalbum .

FROM scratch
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /immich-smartalbum /immich-smartalbum
USER appuser
ENTRYPOINT ["/immich-smartalbum"]
