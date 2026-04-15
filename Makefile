DB_USERNAME = user
DB_PASSWORD := user
ENV := $(shell grep -w 'APP_ENV' .env)
UNAME := $(shell uname -a)
ifeq ($(findstring WSL,"$(UNAME)"),WSL)
	IS_WSL=true
else
	IS_WSL=false
endif
REPO = data-fetcher-api
PORT = 2000

default: up migrate
build: reset
rebuild: reset
reset: env down owner timeout docker
	docker-compose -f docker-compose.yml build
	docker-compose -f docker-compose.yml up -d --remove-orphans
	make package
	make migrate
	make down

cache:
	make console command="src/Command/Cronjob/StorePerformanceInCacheCommand.go"
	make console command="src/Command/Cronjob/StoreTopTradersCommand.go"

config-yaml:
	@vim .env -c "set ff=unix" -c ":wq"
	docker-compose -f docker-compose.yml exec -T api.app bash -c "cd /var/www/html/$(REPO)/ && \
	go run src/Command/General/ConfigConverterCommand.go"

console:
ifeq ($(command),doctrine:migrations:migrate --no-interaction)
	$(MAKE) migrate
else
	docker-compose -f docker-compose.yml exec -T api.app bash -c \
    "cd /var/www/html/$(REPO)/ && go run $$command"
endif

console-async:
ifeq ($(command),mq:subscribe)
	docker-compose -f docker-compose.yml exec -T api.app bash -c \
	"cd /var/www/html/$(REPO)/ && fuser -k -n tcp $(PORT) || true"
	docker-compose -f docker-compose.yml exec -T api.app bash -c \
	"cd /var/www/html/$(REPO)/ && cd /var/www/html/$(REPO)/ && nohup ./data-fetcher-api &> output & sleep 1"
else
	docker-compose -f docker-compose.yml exec -T api.app bash -c \
    	"cd /var/www/html/$(REPO)/ && nohup go run $$command &> console-output & sleep 1"
endif

crontab:
	docker-compose -f docker-compose.yml exec -T api.app bash -c "\
	echo "" > /etc/cron.d/crontab && \
    tail /var/www/deploy/*.cron | tee -a /etc/cron.d/crontab && \
    sed -i '1i PATH="/usr/local/bin:/usr/bin:/bin"' /etc/cron.d/crontab && \
    chmod 0744 /etc/cron.d/crontab && \
    crontab /etc/cron.d/crontab && \
    touch /var/log/cron.log && \
    service cron restart && service cron status"

db-export:
	@echo "Skipped"

db-import:
	make console-async command="src/Command/Cronjob/TradingViewConnectionCommand.go"

docker: timeout
ifeq ($(IS_WSL), true)
	@echo "This is Local"
	rm -rf docker >/dev/null 2>&1
	rm -rf docker.api > /dev/null 2>&1
	git clone --single-branch --branch develop git@bitbucket.org:aitradeai/docker.api.git docker.api
	rm -rf docker/ > /dev/null 2>&1
	mkdir -p docker/
	cp -R docker.api/docker/. docker/
	make owner
	rm -rf docker-compose.yml > /dev/null 2>&1
	cp docker.api/deploy/docker-compose.local.yml docker-compose.yml
	rm -rf .env > /dev/null 2>&1
	cp docker.api/deploy/local.env .env
	rm -rf docker.api > /dev/null 2>&1
else
	@echo "This is Remote"
endif

down:
	docker ps -a -q | xargs -n 1 -P 8 -I {} docker stop {}
	docker builder prune --all --force
	docker system prune -f

encode-password:
	@echo "Skipped"

env:
ifeq ($(IS_WSL), true)
	@echo "This is Local"
	cp deploy/local.env ../$(REPO)/.env > /dev/null 2>&1
else
	@echo "This is Remote"
endif

fixture:
	@echo "Skipped"

init: destroy
destroy: env down timeout docker
	docker volume prune -f
	docker-compose -f docker-compose.yml build --no-cache
	docker-compose -f docker-compose.yml up -d --remove-orphans
	make package
	make migrate
ifneq ($(APP_ENV),'prod')
	make fixture
	make db-import
endif
	make down

jwt:
	@echo "Skipped"

lint: lint-check timeout
lint-check:
	@echo "Skipped"

lint-fix: timeout
	@echo "Skipped"

log: timeout
	docker-compose -f docker-compose.yml ps
	sleep 2
	docker-compose logs -f

migrate:
	docker-compose -f docker-compose.yml exec -T api.app bash -c "cd /var/www/html/$(REPO)/ && \
	go run src/Command/General/MigrateCommand.go"
	make config-yaml

mq-subscribe:
	make console-async command="mq:subscribe"

owner:
    # docker
	mkdir -p docker/mysql/conf/conf.d
	touch docker/mysql/conf/my.cnf
	touch docker/mysql/conf/conf.d/docker.cnf
	touch docker/mysql/conf/conf.d/mysql.cnf
	chmod -R 777 docker
	chmod 644 docker/mysql/conf/my.cnf
	chmod 644 docker/mysql/conf/conf.d/docker.cnf
	chmod 644 docker/mysql/conf/conf.d/mysql.cnf
	# REPO var
	for d in $$(ls ../); do \
	chown -R www-data:www-data ../$$d \
	&& mkdir -p ../$$d/var/cache/$(APP_ENV)/ \
	&& mkdir -p ../$$d/var/log/ \
	&& touch ../$$d/var/log/$(APP_ENV).log \
	&& chmod 777 -R ../$$d/var/ \
	; done
	# REPO public
	mkdir -p ../$(REPO)/public/
	chmod 777 -R ../$(REPO)/public/

package: timeout package1 owner

package1:
	docker-compose -f docker-compose.yml exec -T api.app bash -c \
	"cd /var/www/html/$(REPO)/ && cd /var/www/html/$(REPO)/ && go get"
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.51.2

report:
	git shortlog -s -n -e

serve:
	make cache
	make jwt
	make console command="doctrine:migrations:migrate --no-interaction"
	make crontab
	make mq-subscribe
	make fixture

serve-all:
	for d in $$(ls ../); do cd ../$$d && make serve; done

ssh: timeout
	docker-compose -f docker-compose.yml exec api.app bash

test:
	@echo "Skipped"

test-coverage-compare:
	@echo "Skipped"

timeout:
	export DOCKER_CLIENT_TIMEOUT=2000
	export COMPOSE_HTTP_TIMEOUT=2000

up: docker env down timeout owner
	docker-compose -f docker-compose.yml up -d --remove-orphans
	docker-compose -f docker-compose.yml exec -T api.app bash -c \
	"cd /var/www/html/$(REPO)/ && fuser -k -n tcp $(PORT) || true"
	docker-compose -f docker-compose.yml exec -T api.app bash -c \
	"cd /var/www/html/$(REPO)/ && cd /var/www/html/$(REPO)/ && go run main.go"
