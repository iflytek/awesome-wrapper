FROM 172.16.59.153/aiaas/ubuntu:14.04.gcc.golang
MAINTAINER sqjian@iflytek.com

COPY live /bin/
RUN chmod +x /bin/live
RUN apt-get  update && \
    apt-get -y install gdb  && \
    rm -r /var/lib/apt/lists/*