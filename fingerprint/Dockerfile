FROM smallstep/step-cli

USER root
RUN apk update
RUN apk add inotify-tools
COPY entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
