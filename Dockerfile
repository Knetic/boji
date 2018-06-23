FROM debian:9

WORKDIR /var/lib/boji
EXPOSE 8080

COPY ./.output/*.deb /etc/boji/packages/
RUN dpkg -i /etc/boji/packages/*.deb; \
    apt-get install -f;

CMD ["/usr/local/bin/boji"]