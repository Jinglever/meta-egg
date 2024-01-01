package template

var TplPackageDockerfile = `ARG ci_commit_ref_name
ARG ci_commit_short_sha
ARG pipe_date

FROM golang:1.19-alpine3.15 as build-env

ARG ci_commit_ref_name
ARG ci_commit_short_sha
ENV CI_COMMIT_REF_NAME=$ci_commit_ref_name
ENV CI_COMMIT_SHORT_SHA=$ci_commit_short_sha
ENV PIPE_DATE=$pipe_date

WORKDIR /app

RUN apk add make git gcc g++

COPY . .

RUN /usr/bin/make build
COPY configs/*.yml /app/configs/
RUN rm /app/configs/*-local.yml || true

FROM alpine:3.15

RUN apk add tzdata

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

WORKDIR /app

COPY --from=build-env /app/configs/ /app/configs/
COPY --from=build-env /app/build/bin/* /app/bin/
ENV PATH="/app/bin:${PATH}"
`
