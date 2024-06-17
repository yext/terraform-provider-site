package main

import (
	"github.com/hashicorp/terraform/plugin"

	"github.com/yext/terraform-provider-site/site"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: site.Provider,
	})
}
