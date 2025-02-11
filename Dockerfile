FROM ubuntu:22.10

RUN echo 'hosts: files dns' >> /etc/nsswitch.conf

RUN set -ex && \
    mkdir -p /usr/bin /etc/categraf 

COPY categraf  /usr/bin/categraf

COPY conf /etc/categraf/conf

COPY entrypoint.sh /entrypoint.sh

RUN chmod 755 /entrypoint.sh

CMD ["/entrypoint.sh"]
