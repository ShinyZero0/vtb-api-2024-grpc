FROM smallstep/step-ca

# USER root
# RUN apk update
# RUN apk add tree
# RUN mkdir -p /var/local/step
# RUN chown step:step /var/local/step
USER step
COPY entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
CMD exec /usr/local/bin/step-ca --password-file $PWDPATH $CONFIGPATH
