FROM golang:1.22.3-alpine3.20
# RUN addgroup web &&\
#  adduser -h /home/gin-blog -s /bin/false gb1_web web &&\
#  chmod a+rx /home/gin-blog
#VOLUME /home/gin-blog
# USER gb1_web
# WORKDIR /home/gin-blog
WORKDIR /usr/src/gin-blog
#ENTRYPOINT ["entrypoint.sh"]
CMD ["go", "run", "."]
