FROM --platform=linux/x86-64 golang:1.19.2-alpine3.16 as node-build
RUN apk update && apk add libstdc++ g++ git linux-headers

WORKDIR /src
# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build main.go

FROM --platform=linux/x86-64 alpine:3.14
RUN mkdir -p /arcana_dkg
COPY --from=node-build /src/main /arcana_dkg/dkg


EXPOSE 80 443 1080 8000
VOLUME ["/arcana_dkg"]
CMD ["/arcana_dkg/dkg", "start"]
