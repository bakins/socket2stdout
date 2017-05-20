FROM scratch

COPY ./socket2stdout /
ENTRYPOINT [ "/socket2stdout" ]
