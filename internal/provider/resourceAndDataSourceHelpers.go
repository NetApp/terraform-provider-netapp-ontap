package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
)

type resourceOrDataSourceConfig struct {
	client         *restclient.RestClient
	providerConfig Config
	name           string
}

// getRestClient will use existing client config.client or create one if it's not set
func getRestClient(ctx context.Context, diags diag.Diagnostics, config resourceOrDataSourceConfig, cxProfileName types.String) (*restclient.RestClient, error) {

	if config.client == nil {
		client, err := config.providerConfig.NewClient(ctx, diags, cxProfileName.ValueString(), config.name)
		if err != nil {
			return nil, err
		}
		config.client = client
	}
	return config.client, nil
}
