FROM alpine:3.5
WORKDIR /VaultBacker
ADD VaultBacker /VaultBacker
ADD config.json /VaultBacker
RUN apk --no-cache add ca-certificates && rm -rf /var/cache/apk/* && update-ca-certificates
RUN cd /VaultBacker
ENTRYPOINT ./VaultBacker
EXPOSE 8080