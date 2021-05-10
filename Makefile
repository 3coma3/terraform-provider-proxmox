
.PHONY:  build clean install

all: build

build: clean
	@echo " -> Building"
	@cd cmd/terraform-provider-proxmox && go build
	@echo "Built terraform-provider-proxmox"

install: clean
	@echo " -> Installing"
	go install github.com/3coma3/terraform-provider-proxmox/cmd/terraform-provider-proxmox

clean:
	@git clean -f -d -X
