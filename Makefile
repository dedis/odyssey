.PHONY: cryptutil

all: cothority catadmin cryptutil pcadmin

CO_VER=v3.4.5

cothority:
	@cd /tmp && \
	rm -rf /tmp/cothority && \
	echo "cloning cothority into /tmp..." && \
	git clone https://github.com/dedis/cothority && \
	cd cothority && \
	git checkout $(CO_VER) && \
	go install ./byzcoin/bcadmin && \
	echo "📌 bcadmin $(CO_VER) installed globally" && \
	go install ./calypso/csadmin && \
	echo "📌 csadmin $(CO_VER) install globally" && \
	rm -rf /tmp/cothority

catadmin:
	@cd catalogc/catadmin && \
	go install && \
	echo "📌 catadmin installed globally"

cryptutil:
	@cd cryptutil && \
	go install && \
	echo "📌 cryptutil installed globally"

pcadmin:
	@cd projectc && \
	go install && \
	echo "📌 pcadmin installed globally"

test:
	@echo "🔎 testing cryptutil..." && cd cryptutil && ./test.sh > /dev/null && echo "...✔️ test OK"
	@echo "🔎 testing catalogc..." && cd catalogc && go test ./... > /dev/null && echo "...✔️ test OK"
	@echo "🔎 testing domanager..." && cd domanager/app && go test ./... > /dev/null && echo "...✔️ test OK"
	@echo "🔎 testing dsmanager..." && cd dsmanager/app && go test ./... > /dev/null && echo "...✔️ test OK"
	@echo "🔎 testing enmanager..." && cd enclavem/app && go test ./... > /dev/null && echo "...✔️ test OK"
	@echo "🔎 testing projectc..." && cd projectc && go test ./... > /dev/null && echo "...✔️ test OK"

lint:
	# Coding style static check.
	@go get -v honnef.co/go/tools/cmd/staticcheck
	@go mod tidy
	staticcheck ./...
