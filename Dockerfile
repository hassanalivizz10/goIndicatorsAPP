#first stage - builder
FROM golang:1.21.2 as indicatorsApp
COPY . /indicatorsBuildAPP
WORKDIR /indicatorsBuildAPP
ENV GO111MODULE=on
RUN CGO_ENABLED=0 GOOS=linux go build -o indicatorsBuildAPP
#second stage
FROM alpine:latest
WORKDIR /root/
RUN apk add --no-cache tzdata
COPY --from=indicatorsApp /indicatorsBuildAPP .
CMD ["./indicatorsBuildAPP"]