FROM         scratch
MAINTAINER   Camilo Aguilar <camilo.aguilar@gmail.com>
ARG          NAME
ARG          VERSION
COPY         build/${NAME}_${VERSION}_linux_amd64/${NAME} ${NAME}
EXPOSE       9998 9999
ENTRYPOINT   ["/hello"]
