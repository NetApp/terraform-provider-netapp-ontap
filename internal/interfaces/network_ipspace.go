package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// IPspaceGetDataModelONTAP describes the GET record data model using go types for mapping
type IPspaceGetDataModelONTAP struct {
	Name  string  `mapstructure:"name"`
	UUID  string  `mapstructure:"uuid"`
}

// IPspaceResourceBodyDataModel describes the resource body data model using go types for mapping
type IPspaceResourceBodyDataModel struct {
	Name  *string  `mapstructure:"name"`
}

// IPspacesDataSourceFilterModel describes the filter model for IPspaces
type IPspacesDataSourceFilterModel struct {
	Name  string  `mapstructure:"name"`
}

// GetIPspace to get IPspace info given its name
func GetIPspace(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string) (*IPspaceGetDataModelONTAP, error) {
	api := "network/ipspaces"
	query := r.NewQuery()
	query.Set("name", name)
	query.Fields([]string{
		"name",
		"uuid",
	})

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error reading IPspace info",
			fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP IPspaceGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response),
		)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read ipspace data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetIPspaces to get a list of IPspaces info for given filter options
func GetIPspaces(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string) ([]IPspaceGetDataModelONTAP, error) {
	api := "network/ipspaces"
	query := r.NewQuery()
	if name != "" {
		query.Set("name", name)
	}
	query.Fields([]string{
		"name",
		"uuid",
	})

	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error reading IPspaces info",
			fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP []IPspaceGetDataModelONTAP
	for _, info := range response {
		var record IPspaceGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(
				fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info),
			)
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read ipspaces data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateIPspace to create an IPspace
func CreateIPspace(errorHandler *utils.ErrorHandler, r restclient.RestClient, request IPspaceResourceBodyDataModel) (*IPspaceGetDataModelONTAP, error) {
	api := "network/ipspaces"
	
	var body map[string]interface{}
	if err := mapstructure.Decode(request, &body); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error encoding network/ipspaces body",
			fmt.Sprintf("error on encoding %s body: %s, request: %#v", api, err, request),
		)
	}
	
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, body)
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error creating IPspace",
			fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP IPspaceGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error decoding IPspace info",
			fmt.Sprintf("error on decoding network/ipspaces info: %s, statusCode %d, response %#v", err, statusCode, response),
		)
	}

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Created ipspace resource: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteIPspace to delete an IPspace
func DeleteIPspace(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := fmt.Sprintf("network/ipspaces/%s", uuid)
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError(
			"error deleting IPspace",
			fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode),
		)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Deleted ipspace resource: %s", uuid))
	return nil
}
