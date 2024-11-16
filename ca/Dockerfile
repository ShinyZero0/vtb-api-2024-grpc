FROM smallstep/step-ca

USER step
COPY entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
CMD exec /usr/local/bin/step-ca --password-file $PWDPATH $CONFIGPATH
