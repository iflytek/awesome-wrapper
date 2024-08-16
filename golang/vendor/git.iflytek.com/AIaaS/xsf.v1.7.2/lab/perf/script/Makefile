clientCfg=client.yml
serverCfg=server.yml

.PHONY: run client client-hystrix server

run:
	@echo please specify target!!!

client:
	@docker-compose -f $(clientCfg) down
	@docker-compose -f $(clientCfg) pull
	@docker-compose -f $(clientCfg) up --scale client=1 --scale client-hystrix=0

client-hystrix:
	@docker-compose -f $(clientCfg) down
	@docker-compose -f $(clientCfg) pull
	@docker-compose -f $(clientCfg) up --scale client=0 --scale client-hystrix=1

server:
	@docker-compose -f $(serverCfg) down
	@docker-compose -f $(serverCfg) pull
	@docker-compose -f $(serverCfg) up
