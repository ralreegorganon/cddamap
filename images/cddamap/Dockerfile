FROM alpine:3.7

RUN apk --no-cache add ca-certificates

ADD bin/linux/cddamap /bin/cddamap
ADD /migrations /migrations
ENV CDDAMAP_MIGRATIONS_PATH=file:///migrations

ENTRYPOINT ["cddamap"]