TARGET_DIR = ./dist
VENOM_BIN = $(TARGET_DIR)/venom
COMPOSE_BIN = $(TARGET_DIR)/docker-compose
API_BIN = $(TARGET_DIR)/appcatalog

all: api webui

$(TARGET_DIR):
	$(info Creating $(TARGET_DIR) directory)
	@mkdir -p $(TARGET_DIR)

$(VENOM_BIN): $(TARGET_DIR)
	$(info Installing venom...)
	@curl -L -o $(VENOM_BIN) https://github.com/ovh/venom/releases/download/v0.17.0/venom.linux-amd64
	@chmod +x $(VENOM_BIN)

$(COMPOSE_BIN): $(TARGET_DIR)
	$(info Installing docker-compose...)
	@curl -L https://github.com/docker/compose/releases/download/1.17.0/docker-compose-`uname -s`-`uname -m` -o $(COMPOSE_BIN)
	@chmod +x $(COMPOSE_BIN)

$(API_BIN):
	make -C api server

api:
	make -C api

webui:
	make -C webui

test:
	make -C api test
	make -C webui test

run: all
	./dist/appcatalog

clean:
	make -C api clean
	make -C webui clean

integration-test: $(COMPOSE_BIN) $(VENOM_BIN) $(API_BIN)
	$(COMPOSE_BIN) up -d
	sleep 10;
	{ ./${API_BIN} ${DEBUG} & }; \
	pid=$$!; \
	sleep 5; \
	APP_HOST=http://localhost:8081 $(VENOM_BIN) run --strict --output-dir=$(TARGET_DIR) tests/; \
	r=$$?; \
	kill $$pid; \
	$(COMPOSE_BIN) down; \
	exit $$r

.PHONY: all test run clean integration-test api webui
