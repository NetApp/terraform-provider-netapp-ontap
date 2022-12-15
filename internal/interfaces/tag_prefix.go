package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// TODO:
// copy this file to match you data source (should match internal/interfaces/tag_prefix.go)
// replace GoPrefix with the name of the resource, following go conventions, eg IPInterface
// replace tag_prefix with the name of the resource, for logging purposes, eg ip_interface
// replace api_url with API, eg ip/interfaces
// delete these 5 lines

// GoPrefixGetDataModelONTAP describes the GET record data model using go types for mapping.
type GoPrefixGetDataModelONTAP struct {
	Name string
}

// GetGoPrefix to get tag_prefix info
func GetGoPrefix(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmName string) (*GoPrefixGetDataModelONTAP, error) {
	api := "api_url"
	query := r.NewQuery()
	query.Set("name", name)
	if svmName == "" {
		query.Set("scope", "cluster")
	} else {
		query.Set("svm.name", svmName)
		query.Set("scope", "svm")
	}
	query.Fields([]string{"name", "svm.name", "ip", "scope"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading tag_prefix info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP GoPrefixGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read tag_prefix data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetGoAllPrefix to get tag_prefix info for all resources matching a filter
func GetGoAllPrefix(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *GoPrefixGetDataModelONTAP) ([]GoPrefixGetDataModelONTAP, error) {
	api := "api_url"
	query := r.NewQuery()
	query.Fields([]string{"name", "svm.name", "scope"})
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding tag_prefix filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading tag_prefix info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []GoPrefixGetDataModelONTAP
	for _, info := range response {
		var record GoPrefixGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read tag_prefix data source: %#v", dataONTAP))
	return dataONTAP, nil
}
