FROM golang:1.25.6 AS base

ARG UID=1000
ARG GID=1000

RUN groupadd --system --gid $GID app
RUN useradd --no-log-init --system --create-home --gid $GID --uid $UID app

USER $UID:$GID

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download -x

COPY . .

FROM base AS build

ARG SOURCE_DATE_EPOCH=0
ENV CGO_ENABLED=0

RUN go build -ldflags="-w -s -buildid=" -trimpath -buildvcs=false -o /go/bin/app

FROM base AS dev

RUN go install github.com/air-verse/air@v1.64.0

ENTRYPOINT ["air", "-c", ".air.toml", "--"]

FROM base AS test

ENTRYPOINT ["go", "test"]
CMD ["-v", "-race", "-covermode=atomic", "-coverprofile=/go/src/coverage.out", "./..."]

FROM gcr.io/distroless/static-debian13:nonroot

COPY ./internal/views/pages/html/static /home/nonroot/internal/views/pages/html/static
COPY --from=build /go/bin/app /
ENTRYPOINT ["/app"]
