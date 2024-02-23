package interfaces

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// SvmGetDataModelONTAP describes the GET record data model using go types for mapping.
type SvmGetDataModelONTAP struct {
	Name string
	UUID string
}

// SvmDataModelONTAP describes the svm info required by other API's request.
type SvmDataModelONTAP struct {
	Name string `mapstructure:"name,omitempty"`
	UUID string `mapstructure:"uuid,omitempty"`
}

// SvmResourceModel describes the resource data model.
type SvmResourceModel struct {
	Name           string              `mapstructure:"name,omitempty"`
	Ipspace        Ipspace             `mapstructure:"ipspace"`
	SnapshotPolicy SnapshotPolicy      `mapstructure:"snapshot_policy,omitempty"`
	SubType        string              `mapstructure:"subtype,omitempty"`
	Comment        string              `mapstructure:"comment"`
	Language       string              `mapstructure:"language,omitempty"`
	MaxVolumes     string              `mapstructure:"max_volumes,omitempty"`
	Aggregates     []map[string]string `mapstructure:"aggregates"`
}

// SvmGetDataSourceModel describes the data source model.
type SvmGetDataSourceModel struct {
	Name           string         `mapstructure:"name"`
	UUID           string         `mapstructure:"uuid"`
	Ipspace        Ipspace        `mapstructure:"ipspace"`
	SnapshotPolicy SnapshotPolicy `mapstructure:"snapshot_policy"`
	SubType        string         `mapstructure:"subtype,omitempty"`
	Comment        string         `mapstructure:"comment,omitempty"`
	Language       string         `mapstructure:"language,omitempty"`
	Aggregates     []Aggregate    `mapstructure:"aggregates,omitempty"`
	MaxVolumes     string         `mapstructure:"max_volumes,omitempty"`
}

// Ipspace describes the resource data model.
type Ipspace struct {
	Name string `mapstructure:"name,omitempty"`
}

// SnapshotPolicy describes the resource data model.
type SnapshotPolicy struct {
	Name string `mapstructure:"name,omitempty"`
}

// SvmDataSourceFilterModel describes the data source data model for queries.
type SvmDataSourceFilterModel struct {
	Name string `mapstructure:"name"`
}

// GetSvm to get svm info by uuid
func GetSvm(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) (*SvmGetDataSourceModel, error) {
	statusCode, response, err := r.GetNilOrOneRecord("svm/svms/"+uuid, nil, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading svm info", fmt.Sprintf("error on GET svm/svms: %s, statusCode %d", err, statusCode))
	}

	var dataONTAP *SvmGetDataSourceModel
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("failed to decode response from GET svm", fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read svm info: %#v", dataONTAP))
	return dataONTAP, nil
}

// GetSvmByName to get svm info by name
func GetSvmByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string) (*SvmGetDataSourceModel, error) {
	query := r.NewQuery()
	query.Add("name", name)
	statusCode, response, err := r.GetNilOrOneRecord("svm/svms", query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading svm info", fmt.Sprintf("error on GET svm/svms: %s, statusCode %d", err, statusCode))
	}

	if response == nil {
		tflog.Debug(errorHandler.Ctx, fmt.Sprintf("svm %s not found", name))
		return nil, errorHandler.MakeAndReportError("error reading svm info",
			fmt.Sprintf("svm %s not found", name))
	}

	var dataONTAP *SvmGetDataSourceModel
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("failed to decode response from GET svm by name", fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read svm info: %#v", dataONTAP))
	return dataONTAP, nil
}

// GetSvmByNameIgnoreNotFound to get svm info by name
func GetSvmByNameIgnoreNotFound(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string) (*SvmGetDataSourceModel, error) {
	query := r.NewQuery()
	query.Add("name", name)
	statusCode, response, err := r.GetNilOrOneRecord("svm/svms", query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading svm info", fmt.Sprintf("error on GET svm/svms: %s, statusCode %d", err, statusCode))
	}

	if response == nil {
		return nil, nil
	}

	var dataONTAP *SvmGetDataSourceModel
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("failed to decode response from GET svm by name", fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read svm info: %#v", dataONTAP))
	return dataONTAP, nil
}

// GetSvmByNameDataSource to get data source svm info
func GetSvmByNameDataSource(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string) (*SvmGetDataSourceModel, error) {
	api := "svm/svms"
	query := r.NewQuery()
	query.Fields([]string{"name", "ipspace", "snapshot_policy", "subtype", "comment", "language", "max_volumes", "aggregates"})
	query.Add("name", name)
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading svm info", fmt.Sprintf("error on GET svm/svms: %s, statusCode %d", err, statusCode))
	}

	var dataONTAP SvmGetDataSourceModel
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("failed to decode response from GET svm by name", fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read svm info: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetSvmsByName to get data source list svm info
func GetSvmsByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *SvmDataSourceFilterModel) ([]SvmGetDataSourceModel, error) {
	api := "svm/svms"
	query := r.NewQuery()
	query.Fields([]string{"name", "ipspace", "snapshot_policy", "subtype", "comment", "language", "max_volumes", "aggregates"})

	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding svms filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}

	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading svm info", fmt.Sprintf("error on GET svm/svms: %s, statusCode %d", err, statusCode))
	}

	var dataONTAP []SvmGetDataSourceModel
	for _, info := range response {
		var record SvmGetDataSourceModel
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError("failed to decode response from GET svms by name", fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
		}
		dataONTAP = append(dataONTAP, record)
	}

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read svm info: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateSvm to create svm
func CreateSvm(errorHandler *utils.ErrorHandler, r restclient.RestClient, data SvmResourceModel, setAggrEmpty bool, setCommentEmpty bool) (*SvmGetDataModelONTAP, error) {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding svm body", fmt.Sprintf("error on encoding svm/svms body: %s, body: %#v", err, data))
	}
	if setAggrEmpty {
		delete(body, "aggregates")
	}
	if setCommentEmpty {
		delete(body, "comment")
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod("svm/svms", query, body)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating svm", fmt.Sprintf("error on POST svm/svms: %s, statusCode %d", err, statusCode))

	}

	var dataONTAP SvmGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("failed to decode response from POST svm/svms", fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create svm source - udata: %#v", dataONTAP))
	return &dataONTAP, nil

}

// DeleteSvm to delete svm
func DeleteSvm(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := "svm/svms/" + uuid
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting svm", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))

	}
	return nil
}

// UpdateSvm to update a svm
func UpdateSvm(errorHandler *utils.ErrorHandler, r restclient.RestClient, data SvmResourceModel, uuid string, setAggrEmpty bool, setCommentEmpty bool) error {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return errorHandler.MakeAndReportError("error encoding svm body", fmt.Sprintf("error on encoding svm/svms body: %s, body: %#v", err, data))
	}
	// delete body if there is no change - comment can be changed to empty, aggregate can be changed to empty
	if !setAggrEmpty && len(data.Aggregates) == 0 {
		delete(body, "aggregates")
	}
	if !setCommentEmpty && data.Comment == "" {
		delete(body, "comment")
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Update svm info: %#v", data))
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, _, err := r.CallUpdateMethod("svm/svms/"+uuid, query, body)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating svm", fmt.Sprintf("error on PATCH svm/svms: %s, statusCode %d", err, statusCode))
	}
	return nil
}

// ValidateIntORString to validate int or string
func ValidateIntORString(errorHandler *utils.ErrorHandler, value string, astring string) error {
	if value == "" || value == astring {
		return nil
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return errorHandler.MakeAndReportError("falied to validate", fmt.Sprintf("Error: expecting int value or '%s', got: %s", astring, value))
	}

	stringValue := strconv.Itoa(intValue)
	if stringValue != value {
		return errorHandler.MakeAndReportError("falied to validate", fmt.Sprintf("Error: expecting int value or '%s', got: %s", astring, value))
	}
	return nil
}
