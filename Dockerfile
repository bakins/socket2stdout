FROM scratch

COPY ./socket2stdout.linux.amd64 /socket2stdout
ENTRYPOINT [ "/socket2stdout" ]
