package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// IPServicePolicyGetDataModelONTAP describes the GET record data model
type IPServicePolicyGetDataModelONTAP struct {
	SVM      svm      `mapstructure:"svm"`
	Name     string   `mapstructure:"name"`
	Scope    string   `mapstructure:"scope"`
	IPspace  *IPspaceGetDataModel   `mapstructure:"ipspace"`
	Services []string `mapstructure:"services"`
	UUID     string   `mapstructure:"uuid"`
}

// IPspaceGetDataModel describes the GET record data model using go types for mapping
type IPspaceGetDataModel struct {
	Name string `mapstructure:"name,omitempty"`
}

// IPServicePolicyResourceBodyDataModel describes the resource body data model
type IPServicePolicyResourceBodyDataModel struct {
	Name     string   `mapstructure:"name,omitempty"`
	SVM      svm      `mapstructure:"svm,omitempty"`
	Scope    string   `mapstructure:"scope,omitempty"`
	IPspace  string   `mapstructure:"ipspace,omitempty"`
	Services []string `mapstructure:"services"`
}

// IPServicePolicyDataSourceFilterModel describes the filter model for service policies
type IPServicePolicyDataSourceFilterModel struct {
	Scope   string `mapstructure:"scope"`
	SVMName string `mapstructure:"svm.name"`
	Name    string `mapstructure:"name"`
}

// GetServicePolicyByUUID to get service policy info given its UUID
func GetServicePolicyByUUID(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string, version versionModelONTAP) (*IPServicePolicyGetDataModelONTAP, error) {
	if version.Generation == 9 && version.Major < 8 {
		return nil, errorHandler.MakeAndReportError("error reading service policy info", "network/ip/service-policies API supported on ONTAP version 9.8 or later")
	}
	api := fmt.Sprintf("network/ip/service-policies/%s", uuid)
	query := r.NewQuery()
	query.Fields([]string{
		"name",
		"svm",
		"uuid",
		"scope",
		"ipspace",
		"services",
	})

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading service policy info", fmt.Sprintf("error on GET %s: %s", api, err))
	}
	var dataONTAP IPServicePolicyGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding service policy info", fmt.Sprintf("error on decode %s: %s, statusCode %d, response %#v", api, err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read ip_service_policy data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetServicePolicyByName to get service policy info given its name, svm, ipspace
func GetServicePolicyByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmName string, ipspace string, version versionModelONTAP) (*IPServicePolicyGetDataModelONTAP, error) {
	if version.Generation == 9 && version.Major < 8 {
		return nil, errorHandler.MakeAndReportError("error reading service policy info", "network/ip/service-policies API supported on ONTAP version 9.8 or later")
	}
	
	api := "network/ip/service-policies"
	query := r.NewQuery()
	query.Set("name", name)
	if svmName == "" {
		query.Set("scope", "cluster")
	} else {
		query.Set("svm.name", svmName)
		query.Set("scope", "svm")
	}
	if ipspace != "" {
		query.Set("ipspace.name", ipspace)
	}
	query.Fields([]string{
		"name",
		"svm",
		"uuid",
		"scope",
		"ipspace",
		"services",
	})

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading service policy info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP IPServicePolicyGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read ip_service_policy data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetServicePolicies to get list of service policies info for given filter options
func GetServicePolicies(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *IPServicePolicyDataSourceFilterModel, version versionModelONTAP) ([]IPServicePolicyGetDataModelONTAP, error) {
	if version.Generation == 9 && version.Major < 8 {
		return nil, errorHandler.MakeAndReportError("error reading service policy info", "network/ip/service-policies API supported on ONTAP version 9.8 or later")
	}
	
	api := "network/ip/service-policies"
	query := r.NewQuery()
	query.Fields([]string{
		"name",
		"svm",
		"uuid",
		"scope",
		"ipspace",
		"services",
	})

	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding service policy filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading service policy info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []IPServicePolicyGetDataModelONTAP
	for _, info := range response {
		var record IPServicePolicyGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read ip_service_policies data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateIPServicePolicy to create a service policy
func CreateIPServicePolicy(errorHandler *utils.ErrorHandler, r restclient.RestClient, request IPServicePolicyResourceBodyDataModel) (*IPServicePolicyGetDataModelONTAP, error) {
	api := "network/ip/service-policies"
	var body map[string]interface{}
	if err := mapstructure.Decode(request, &body); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding network/ip/service-policies body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, body)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating service policy", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP IPServicePolicyGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding network/ip/service-policies info", fmt.Sprintf("error on decode network/ip/service-policies info: %s, statusCode %d, response %#v", err, statusCode, response))
	}

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Created service policy source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteIPServicePolicy to delete a service policy
func DeleteIPServicePolicy(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := fmt.Sprintf("network/ip/service-policies/%s", uuid)
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting service policy", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}

// UpdateIPServicePolicy to modify a service policy
func UpdateIPServicePolicy(errorHandler *utils.ErrorHandler, r restclient.RestClient, request IPServicePolicyResourceBodyDataModel, uuid string) error {
	api := fmt.Sprintf("network/ip/service-policies/%s", uuid)
	var body map[string]interface{}
	if err := mapstructure.Decode(request, &body); err != nil {
		return errorHandler.MakeAndReportError("error encoding network/ip/service-policies body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, _, err := r.CallUpdateMethod(api, query, body)
	if err != nil {
		return errorHandler.MakeAndReportError("error modifying service policy", fmt.Sprintf("error on PATCH %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
