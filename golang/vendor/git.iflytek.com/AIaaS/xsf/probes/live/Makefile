SHELL:=/bin/bash

.PHONY:run exec test linux

export GODEBUG=""

containerName="fuck"

run:
	docker pull 172.16.59.153/aiaas/xsf:2.0.0
	docker run \
	-it \
	--rm \
	--network=host \
	--name=$containerName \
	-w "/xsf/basic/server" \
	172.16.59.153/aiaas/xsf:2.0.0 \
	./server -m 0 -c server.toml -p 3s -g 3s -s xsf-server -u "http://10.1.87.69:6868"

exec:
	docker cp live $containerName:/bin
	docker exec -it $containerName live -project=guiderAllService -group=gas -service=live -companionUrl=http://10.1.87.69:6868
