FROM smallstep/step-cli

USER root
ADD crontab /var/spool/cron/crontabs/root
RUN chmod 0644 /var/spool/cron/crontabs/root
RUN apk update
RUN apk add inotify-tools jq

COPY entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
# CMD ["/usr/sbin/crond", "-l", "2", "-f"]
CMD sleep 1000000
