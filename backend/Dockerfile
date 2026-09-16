# syntax=docker/dockerfile:1

FROM ubuntu:latest

WORKDIR /app

RUN apt update && apt install tzdata -y
ENV TZ="Europe/Paris"

RUN apt-get update && apt-get install -y ca-certificates
RUN mkdir data
COPY ./.env ./bin/linux_amd64/api ./

EXPOSE 4354

ENTRYPOINT ["/app/api"]
