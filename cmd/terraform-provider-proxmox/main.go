package main

import (
	"github.com/3coma3/terraform-provider-proxmox/proxmox"
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return proxmox.Provider()
		},
	})
}
