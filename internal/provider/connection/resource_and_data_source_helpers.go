package connection

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

type ResourceOrDataSourceConfig struct {
	Client         *restclient.RestClient
	ProviderConfig Config
	Name           string
}

// getRestClient will use existing client config.client or create one if it's not set
func GetRestClient(errorHandler *utils.ErrorHandler, config ResourceOrDataSourceConfig, cxProfileName types.String) (*restclient.RestClient, error) {

	if config.Client == nil {
		client, err := config.ProviderConfig.NewClient(errorHandler, cxProfileName.ValueString(), config.Name)
		if err != nil {
			return nil, err
		}
		config.Client = client
	}
	return config.Client, nil
}

// func flattenTypesInt64List(clist []int64) interface{} {
func FlattenTypesInt64List(clist []int64) []types.Int64 {
	if len(clist) == 0 {
		return nil
	}
	cronUnits := make([]types.Int64, len(clist))
	for index, record := range clist {
		cronUnits[index] = types.Int64Value(record)
	}

	return cronUnits
}

// func FlattenTypesStringList(terraformStringsList []string) interface{} {
func FlattenTypesStringList(terraformStringsList []string) []types.String {
	if len(terraformStringsList) == 0 {
		return nil
	}
	stringsList := make([]types.String, len(terraformStringsList))
	for index, record := range terraformStringsList {
		stringsList[index] = types.StringValue(record)
	}

	return stringsList
}
