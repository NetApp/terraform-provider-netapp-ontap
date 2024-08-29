package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// QOSPoliciesGetDataModelONTAP describes the GET record data model using go types for mapping.
type QOSPoliciesGetDataModelONTAP struct {
	Name     string   `mapstructure:"name"`
	SVM      svm      `mapstructure:"svm"`
	Scope    string   `mapstructure:"scope,omitempty"`
	Fixed    Fixed    `mapstructure:"fixed,omitempty"`
	Adaptive Adaptive `mapstructure:"adaptive,omitempty"`
	UUID     string   `mapstructure:"uuid"`
}

// QOSPoliciesResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type QOSPoliciesResourceBodyDataModelONTAP struct {
	Name     string   `mapstructure:"name"`
	SVM      svm      `mapstructure:"svm"`
	Scope    string   `mapstructure:"scope,omitempty"`
	Fixed    Fixed    `mapstructure:"fixed,omitempty"`
	Adaptive Adaptive `mapstructure:"adaptive,omitempty"`
}

// QOSPoliciesUpdateResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type QOSPoliciesUpdateResourceBodyDataModelONTAP struct {
	Name     string   `mapstructure:"name,omitempty"`
	Fixed    Fixed    `mapstructure:"fixed,omitempty"`
	Adaptive Adaptive `mapstructure:"adaptive,omitempty"`
}

// QOSPoliciesDataSourceFilterModel describes the data source data model for queries.
type QOSPoliciesDataSourceFilterModel struct {
	Name    string `mapstructure:"name"`
	SVMName string `mapstructure:"svm.name"`
}

// Fixed describes the data model for fixed qos policies.
type Fixed struct {
	MaxThroughputIOPS int  `mapstructure:"max_throughput_iops,omitempty"`
	MinThroughputIOPS int  `mapstructure:"min_throughput_iops,omitempty"`
	MaxThroughputMBPS int  `mapstructure:"max_throughput_mbps,omitempty"`
	MinThroughputMBPS int  `mapstructure:"min_throughput_mbps,omitempty"`
	CapacityShared    bool `mapstructure:"capacity_shared,omitempty"`
}

// Adaptive describes the data model for adaptive qos policies.
type Adaptive struct {
	ExpectedIOPSAllocation string `mapstructure:"expected_iops_allocation,omitempty"`
	ExpectedIOPS           int    `mapstructure:"expected_iops,omitempty"`
	PeakIOPSAllocation     string `mapstructure:"peak_iops_allocation,omitempty"`
	BlockSize              string `mapstructure:"block_size,omitempty"`
	PeakIOPS               int    `mapstructure:"peak_iops,omitempty"`
	AbsoluteMinIOPS        int    `mapstructure:"absolute_min_iops,omitempty"`
}

// GetQOSPoliciesByName to get qos_policies info
func GetQOSPoliciesByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmName string) (*QOSPoliciesGetDataModelONTAP, error) {
	api := "storage/qos/policies"
	query := r.NewQuery()
	query.Set("name", name)
	query.Fields([]string{"name", "svm.name", "scope", "fixed", "adaptive", "uuid"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading qos_policies info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP QOSPoliciesGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read qos_policies data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetQOSPoliciesByUUID to get qos_policies info
func GetQOSPoliciesByUUID(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) (*QOSPoliciesGetDataModelONTAP, error) {
	api := "storage/qos/policies/" + uuid
	query := r.NewQuery()
	query.Fields([]string{"name", "svm.name", "scope", "fixed", "adaptive", "uuid"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading qos_policies info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP QOSPoliciesGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read qos_policies data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetQOSPolicies to get qos_policies info for all resources matching a filter
func GetQOSPolicies(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *QOSPoliciesDataSourceFilterModel) ([]QOSPoliciesGetDataModelONTAP, error) {
	api := "storage/qos/policies"
	query := r.NewQuery()
	query.Fields([]string{"name", "svm.name", "scope", "fixed", "adaptive", "uuid"})
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding qos_policies filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading qos_policies info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []QOSPoliciesGetDataModelONTAP
	for _, info := range response {
		var record QOSPoliciesGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read qos_policies data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateQOSPolicies to create qos_policies
func CreateQOSPolicies(errorHandler *utils.ErrorHandler, r restclient.RestClient, body QOSPoliciesResourceBodyDataModelONTAP) (*QOSPoliciesGetDataModelONTAP, error) {
	api := "storage/qos/policies"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding qos_policies body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating qos_policies", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP QOSPoliciesGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding qos_policies info", fmt.Sprintf("error on decode storage/qos_policiess info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create qos_policies source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// UpdateQOSPolicies to update qos_policies
func UpdateQOSPolicies(errorHandler *utils.ErrorHandler, r restclient.RestClient, data QOSPoliciesUpdateResourceBodyDataModelONTAP, id string) error {
	api := "storage/qos/policies"
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return errorHandler.MakeAndReportError("error encoding qos_policies body", fmt.Sprintf("error on encoding qos_policies body: %s, body: %#v", err, data))
	}

	statusCode, _, err := r.CallUpdateMethod(api+"/"+id, nil, body)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating qos_policies", fmt.Sprintf("error on PATCH storage/qos/policies: %s, statusCode %d", err, statusCode))
	}
	return nil
}

// DeleteQOSPolicies to delete qos_policies
func DeleteQOSPolicies(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := "storage/qos/policies"
	statusCode, _, err := r.CallDeleteMethod(api+"/"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting qos_policies", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
