FROM 172.16.59.153/aiaas/ubuntugo:1.9.2
MAINTAINER jbchen5@iflytek.com
COPY bin/ /lb/
RUN chmod 755 /lb/*
WORKDIR /lb
CMD ./lb -m 1 -c lb.toml -p AIaaS -s iat_lb -u http://10.1.86.223:9080 -g aitest   