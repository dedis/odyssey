.PHONY: cryptutil

all: cothority catadmin cryptutil pcadmin

cothority:
	@cd /tmp && \
	rm -rf /tmp/cothority && \
	echo "cloning cothority into /tmp..." && \
	git clone https://github.com/dedis/cothority && \
	cd cothority && \
	git checkout tags/v3.4.4 && \
	go install ./byzcoin/bcadmin && \
	echo "📌 bcadmin v3.4.4 installed globally" && \
	go install ./calypso/csadmin && \
	echo "📌 csadmin v3.4.4 install globally" && \
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
	@echo "🔎 testing projectc..." && cd projectc && go test ./... > /dev/null && echo "...✔️ test OK"