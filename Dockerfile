FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.20-alpine3.17 AS builder
RUN mkdir /app
ADD . /app
WORKDIR /app

RUN apk add --no-cache git

ARG GOPROXY
# download deps before gobuild
RUN go mod download -x

ARG TARGETOS
ARG TARGETARCH
RUN source ./scripts/version.sh && \
  GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -v -o 2mqtt -ldflags "$LD_FLAGS" cmd/main.go

FROM alpine:3.17

LABEL maintainer="Jeeva Kandasamy <jkandasa@gmail.com>"

ENV APP_HOME="/app" \
    DATA_HOME="/mc_home"

# install timzone utils
RUN apk --no-cache add tzdata

# create a user and give permission for the locations
RUN mkdir -p ${APP_HOME} && mkdir -p ${DATA_HOME}

# copy application bin file
COPY --from=builder /app/2mqtt ${APP_HOME}/2mqtt

RUN chmod +x ${APP_HOME}/2mqtt

# copy default files
COPY ./resources/sample-config.yaml ${APP_HOME}/config.yaml

WORKDIR ${APP_HOME}

CMD ["/app/2mqtt", "--config", "/app/config.yaml"]
