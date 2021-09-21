FROM golang:alpine AS backend

RUN apk update && apk add --no-cache git gcc musl-dev
WORKDIR /src
COPY go.mod /src/go.mod
COPY go.sum /src/go.sum
RUN go mod download

COPY . /src
RUN mkdir -p /out
RUN go test ./...
RUN go build ./cmd/moex-bond-recommender/
RUN mv ./moex-bond-recommender /out
RUN cp -R ./templates /out/
RUN cp -R ./www /out/

FROM alpine:latest
WORKDIR /app
COPY --from=backend /out /app
ENTRYPOINT [ "/app/moex-bond-recommender" ]
