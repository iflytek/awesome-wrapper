#!/usr/bin/env bash

docker run -it -w /xsf/basic/server --name=xxx --net=host 172.16.59.153/aiaas/xsf:0.0.0 ./server -m 0 -c server.toml -p 3s -g 3s -s xsf-server -u "http://10.1.87.69:6868"