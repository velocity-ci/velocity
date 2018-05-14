test:
	cd ../ && $(MAKE) dep-install
	docker-compose -f docker-compose.integration.yaml down
	docker-compose -f docker-compose.integration.yaml up --abort-on-container-exit