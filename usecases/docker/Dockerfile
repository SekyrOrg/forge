FROM alpine:latest
#Download the binary from the github release page
#https://github.com/SekyrOrg/forge/releases/latest/download/forge_Linux_x86_64
RUN wget -O forge https://github.com/SekyrOrg/forge/releases/latest/download/forge_Linux_x86_64
RUN chmod +x forge
RUN apk add --no-cache file
RUN file ./forge
RUN ls /usr/bin
RUN ./forge -D \
    -f /usr/bin/md5sum \
    -f /usr/bin/sha256sum  \
    -f /usr/bin/nc #,/bin/sh,/usr/bin/crontab,/sbin/ip,/usr/bin/wget,/sbin/apk,/bin/base64,/bin/chmod,/bin/echo,/bin/hostname,/bin/pwd,/bin/sed,/bin/tar,/usr/bin/id,/usr/bin/awk,/usr/bin/whoami,/usr/bin/find
RUN md5sum forge
