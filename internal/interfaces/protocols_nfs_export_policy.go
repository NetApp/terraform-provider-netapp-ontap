package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// ExportpolicyResourceModel describes the resource data model.
type ExportpolicyResourceModel struct {
	Name string            `mapstructure:"name"`
	Svm  SvmDataModelONTAP `mapstructure:"svm"`
	ID   int               `mapstructure:"id"`
}

// ExportpolicyResourceBodyDataModelONTAP describes the resource data model.
type ExportpolicyResourceBodyDataModelONTAP struct {
	Name string            `mapstructure:"name"`
	Svm  SvmDataModelONTAP `mapstructure:"svm"`
}

// ExportPolicyGetDataModelONTAP describes the GET record data model using go types for mapping.
type ExportPolicyGetDataModelONTAP struct {
	Name string `mapstructure:"name"`
	Svm  string `mapstructure:"svm_name"`
	ID   int    `mapstructure:"id"`
}

// ExportPolicyGetDataFilterModel describes filter model
type ExportPolicyGetDataFilterModel struct {
	Name    string `mapstructure:"name"`
	SVMName string `mapstructure:"svm.name"`
}

// CreateExportPolicy to create export policy
func CreateExportPolicy(errorHandler *utils.ErrorHandler, r restclient.RestClient, data ExportpolicyResourceBodyDataModelONTAP) (*ExportPolicyGetDataModelONTAP, error) {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding export policy body", fmt.Sprintf("error on encoding export policy body: %s, body: %#v", err, data))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod("protocols/nfs/export-policies", query, body)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating export policy", fmt.Sprintf("error on POST protocols/nfs/export-policies: %s, statusCode %d", err, statusCode))
	}

	var dataONTAP ExportPolicyGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding export policies info", fmt.Sprintf("error on decode protocols/nfs/export-policies info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create export policy source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetExportPolicy to get export policy
func GetExportPolicy(errorHandler *utils.ErrorHandler, r restclient.RestClient, id string) (*ExportPolicyGetDataModelONTAP, error) {
	api := "protocols/nfs/export-policies/" + id
	statusCode, response, err := r.GetNilOrOneRecord(api, nil, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading export policy info", fmt.Sprintf("error on GET protocols/nfs/export-policies/%s: %s", id, err))
	}

	var dataONTAP ExportPolicyGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding export policy info", fmt.Sprintf("error on decode protocols/nfs/export-policies/%s: %s, statusCode %d, response %#v", id, err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read export policy source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetNfsExportPolicyByName to get export policy by filter
func GetNfsExportPolicyByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter interface{}) (*ExportPolicyGetDataModelONTAP, error) {
	query := r.NewQuery()
	query.Fields([]string{"name"})
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding ip_interface filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetNilOrOneRecord("protocols/nfs/export-policies", query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading export policy info", fmt.Sprintf("error on GET protocols/nfs/export-policies: %s", err))
	}

	var dataONTAP ExportPolicyGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding export policy info", fmt.Sprintf("error on decode protocols/nfs/export-policies: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read export policy source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetExportPoliciesList to get export policies
func GetExportPoliciesList(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *ExportPolicyGetDataFilterModel) ([]ExportpolicyResourceModel, error) {
	api := "protocols/nfs/export-policies"
	query := r.NewQuery()
	query.Fields([]string{"name", "id", "svm.name", "svm.uuid"})
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding export policies filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading export policies info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []ExportpolicyResourceModel
	for _, info := range response {
		var record ExportpolicyResourceModel
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read export policies data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// DeleteExportPolicy to delete export policy
func DeleteExportPolicy(errorHandler *utils.ErrorHandler, r restclient.RestClient, id string) error {
	statusCode, _, err := r.CallDeleteMethod("protocols/nfs/export-policies/"+id, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting export policy", fmt.Sprintf("error on DELETE protocols/nfs/export-policies/%s: %s, statusCode %d", id, err, statusCode))
	}
	return nil
}

// UpdateExportPolicy updates export policy
func UpdateExportPolicy(errorHandler *utils.ErrorHandler, r restclient.RestClient, data ExportpolicyResourceBodyDataModelONTAP, id string) error {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return errorHandler.MakeAndReportError("error encoding export policy body", fmt.Sprintf("error on encoding export policy body: %s, body: %#v", err, data))
	}
	// svm is not allowed in the API body.
	delete(body, "svm")
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, _, err := r.CallUpdateMethod("protocols/nfs/export-policies/"+id, query, body)
	if err != nil {
		return errorHandler.MakeAndReportError("error creating export policy", fmt.Sprintf("error on POST protocols/nfs/export-policies: %s, statusCode %d", err, statusCode))
	}
	return nil
}
