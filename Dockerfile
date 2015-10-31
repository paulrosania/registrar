FROM debian:jessie

COPY registrar.ini.sample /etc/registrar/registrar.ini
COPY bin/registrar /registrar/bin/registrar

CMD ["/registrar/bin/registrar"]

EXPOSE 80
