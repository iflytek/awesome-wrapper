export GOPROXY=https://goproxy.cn,direct
export GO111MODULE=on
export GOSUMDB=off

.PHONY: all

all: build_example_client build_example_server build_cgroup build_cgroup build_rate_server probes_live build_perf_client build_perf_server build_grpc_client build_grpc_server build_basic_client build_basic_host_adapter build_basic_cmdClient build_basic_server
build_example_client:
	cd example/client && go build -v -mod=vendor
build_example_server:
	cd example/server && go build -v -mod=vendor
build_cgroup:
	cd lab/cgroup && go build -v -mod=vendor
build_rate_client:
	cd lab/rate/client && go build -v -mod=vendor
build_rate_server:
	cd lab/rate/server && go build -v -mod=vendor
probes_live:
	cd probes/live && go build -v -mod=vendor
build_perf_client:
	cd lab/perf/client && go build -v -mod=vendor
build_perf_server:
	cd lab/perf/server && go build -v -mod=vendor
build_grpc_client:
	cd lab/grpc/client && go build -v -mod=vendor
build_grpc_server:
	cd lab/grpc/server && go build -v -mod=vendor
build_basic_client:
	cd lab/basic/client && go build -v -mod=vendor
build_basic_host_adapter:
	cd lab/basic/other/hostAdapter && go build -v -mod=vendor
build_basic_cmdClient:
	cd lab/basic/cmdClient && go build -v -mod=vendor
build_basic_server:
	cd lab/basic/server && go build -v -mod=vendor