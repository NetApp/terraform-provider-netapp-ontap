package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// IPRouteGetDataModelONTAP describes the GET record data model using go types for mapping.
type IPRouteGetDataModelONTAP struct {
	Destination DestinationDataSourceModel `mapstructure:"destination"`
	UUID        string                     `mapstructure:"uuid"`
	Gateway     string                     `mapstructure:"gateway"`
	Metric      int64                      `mapstructure:"metric"`
	SVMName     Vserver                    `mapstructure:"svm"`
}

// IPRouteResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type IPRouteResourceBodyDataModelONTAP struct {
	Destination DestinationDataSourceModel `mapstructure:"destination"`
	SVM         Vserver                    `mapstructure:"svm"`
}

// DestinationDataSourceModel describes the GET record data model using go types for mapping.
type DestinationDataSourceModel struct {
	Address string `mapstructure:"address"`
	Netmask string `mapstructure:"netmask"`
}

// GetIPRoute to get net_route info
func GetIPRoute(errorHandler *utils.ErrorHandler, r restclient.RestClient, Destination string, svmName string, version versionModelONTAP) (*IPRouteGetDataModelONTAP, error) {
	api := "/network/ip/routes"
	query := r.NewQuery()
	query.Set("destination.address", Destination)
	if svmName == "" {
		query.Set("scope", "cluster")
	} else {
		query.Set("svm.name", svmName)
		query.Set("scope", "svm")
	}
	var fields = []string{"destination", "svm.name", "gateway", "scope"}
	if version.Generation == 9 && version.Major > 10 {
		fields = append(fields, "metric")
	}
	query.Fields(fields)
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading /network/ip/routes info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP IPRouteGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read /network/ip/routes data source: %#v", dataONTAP))
	return &dataONTAP, nil
}
