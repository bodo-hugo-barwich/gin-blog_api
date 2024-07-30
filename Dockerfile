FROM golang:1.22.3-alpine3.20
# RUN addgroup web &&\
#  adduser -h /home/cxcurrency -s /bin/false cxc1_web web &&\
#  chmod a+rx /home/cxcurrency
#VOLUME /home/cxcurrency
# USER cxc1_web
# WORKDIR /home/cxcurrency
WORKDIR /usr/src/cxcurrency
#ENTRYPOINT ["entrypoint.sh"]
CMD ["go", "run", "."]
