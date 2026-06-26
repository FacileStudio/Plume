FROM golang:1.25-alpine AS build

ARG TARGETOS=linux
ARG TARGETARCH

WORKDIR /repo/apps/api

COPY apps/api/go.mod apps/api/go.sum ./

RUN go mod download

COPY apps/api ./

ARG CACHE_BUST=1
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH:-amd64} go build -o bin/api .

FROM build AS dirs
RUN mkdir -p /data/uploads/avatars

FROM gcr.io/distroless/static-debian12

COPY --from=dirs /data/uploads /data/uploads
COPY --from=build /repo/apps/api/bin/api /api

EXPOSE 4000

ENTRYPOINT ["/api"]
