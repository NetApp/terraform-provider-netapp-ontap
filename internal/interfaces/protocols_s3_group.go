package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// ProtocolsS3GroupGetDataModelONTAP describes the GET record data model
type ProtocolsS3GroupGetDataModelONTAP struct {
	SVM      svm                  `mapstructure:"svm"`
	Name     string               `mapstructure:"name"`
	Comment  string               `mapstructure:"comment"`
	Users    []UserGetDataModel   `mapstructure:"users"`
	Policies []PolicyGetDataModel `mapstructure:"policies"`
	ID       int64                `mapstructure:"id"`
}

// ProtocolsS3GroupResourceBodyDataModel describes the resource body data model
type ProtocolsS3GroupResourceBodyDataModel struct {
	Name     string                    `mapstructure:"name"`
	Comment  *string                   `mapstructure:"comment,omitempty"`
	Users    []map[string]interface{}  `mapstructure:"users"`
	Policies *[]map[string]interface{} `mapstructure:"policies,omitempty"`
}

// UserGetDataModel describes the GET record data model using go types for mapping
type UserGetDataModel struct {
	Name string `mapstructure:"name"`
}

// PolicyGetDataModel describes the GET record data model using go types for mapping
type PolicyGetDataModel struct {
	Name string `mapstructure:"name"`
}

// S3GroupsFilterModel describes the filter model for S3 groups
type S3GroupsFilterModel struct {
	SVMName string `mapstructure:"svm.name"`
	Name    string `mapstructure:"name"`
}

// GetProtocolsS3GroupbyID to get S3 group info given its ID
func GetProtocolsS3GroupbyID(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmUUID string, id int64, version versionModelONTAP) (*ProtocolsS3GroupGetDataModelONTAP, error) {
	if version.Generation == 9 && version.Major < 8 {
		return nil, errorHandler.MakeAndReportError("error reading S3 group info", "protocols/s3/services/{svm.uuid}/groups API supported on ONTAP version 9.8 or later")
	}
	api := fmt.Sprintf("protocols/s3/services/%s/groups/%d", svmUUID, id)
	statusCode, response, err := r.GetNilOrOneRecord(api, nil, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading S3 Group info", fmt.Sprintf("error on GET %s: %s", api, err))
	}
	var dataONTAP ProtocolsS3GroupGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding S3 Group info", fmt.Sprintf("error on decode %s: %s, statusCode %d, response %#v", api, err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read s3_group data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetProtocolsS3Group to get S3 group info given its name
func GetProtocolsS3Group(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmUUID string, version versionModelONTAP) (*ProtocolsS3GroupGetDataModelONTAP, error) {
	if version.Generation == 9 && version.Major < 8 {
		return nil, errorHandler.MakeAndReportError("error reading S3 group info", "protocols/s3/services/{svm.uuid}/groups API supported on ONTAP version 9.8 or later")
	}
	
	api := fmt.Sprintf("protocols/s3/services/%s/groups", svmUUID)
	query := r.NewQuery()
	query.Set("name", name)
	var fields = []string{"svm", "name", "id", "comment", "users.name", "policies.name"}
	query.Fields(fields)

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading S3 group info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ProtocolsS3GroupGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read s3_group data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetProtocolsS3Groups to get list of S3 groups info for given SVM
func GetProtocolsS3Groups(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmUUID string, version versionModelONTAP, groupName string) ([]ProtocolsS3GroupGetDataModelONTAP, error) {
	if version.Generation == 9 && version.Major < 8 {
		return nil, errorHandler.MakeAndReportError("error reading S3 groups info", "protocols/s3/services/{svm.uuid}/groups API supported on ONTAP version 9.8 or later")
	}
	
	api := fmt.Sprintf("protocols/s3/services/%s/groups", svmUUID)
	query := r.NewQuery()
	if groupName != "" {
		query.Set("name", groupName)
	}
	var fields = []string{"svm", "name", "id", "comment", "users.name", "policies.name"}
	query.Fields(fields)
	
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading S3 groups info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []ProtocolsS3GroupGetDataModelONTAP
	for _, info := range response {
		var record ProtocolsS3GroupGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read s3_groups data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateProtocolS3Group to create a S3 group
func CreateProtocolS3Group(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmUUID string, request ProtocolsS3GroupResourceBodyDataModel) (*ProtocolsS3GroupGetDataModelONTAP, error) {
	api := fmt.Sprintf("protocols/s3/services/%s/groups", svmUUID)
	var body map[string]interface{}
	if err := mapstructure.Decode(request, &body); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding protocols/s3/services/{svm.uuid}/groups body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, body)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating S3 group", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ProtocolsS3GroupGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding protocols/s3/services/{svm.uuid}/groups info", fmt.Sprintf("error on decode protocols/s3/services/{svm.uuid}/groups info: %s, statusCode %d, response %#v", err, statusCode, response))
	}

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create s3_group source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteProtocolsS3Group to delete a S3 group
func DeleteProtocolsS3Group(errorHandler *utils.ErrorHandler, r restclient.RestClient, id int64, svmUUID string) error {
	api := fmt.Sprintf("protocols/s3/services/%s/groups/%d", svmUUID, id)
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting S3 group", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}

// UpdateProtocolsS3Group to modify a S3 group
func UpdateProtocolsS3Group(errorHandler *utils.ErrorHandler, r restclient.RestClient, request ProtocolsS3GroupResourceBodyDataModel, id int64, svmUUID string) error {
	api := fmt.Sprintf("protocols/s3/services/%s/groups/%d", svmUUID, id)
	var body map[string]interface{}
	if err := mapstructure.Decode(request, &body); err != nil {
		return errorHandler.MakeAndReportError("error encoding protocols/s3/services/{svm.uuid}/groups body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, _, err := r.CallUpdateMethod(api, query, body)
	if err != nil {
		return errorHandler.MakeAndReportError("error modifying S3 group", fmt.Sprintf("error on PATCH %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
