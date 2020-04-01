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
	cd cryptutil && ./test.sh