#!/usr/bin/env bash

docker run --rm --net=host -w "/xsf/perf/client" -e "XSF-DEBUG=1" 172.16.59.153/aiaas/xsf:0.0.0 ./client -s 0 -m 1 -pre 0 -tm 2 -goroutines 100 -count 10000
