package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// ProtocolsS3UserGetDataModelONTAP describes the GET record data model using go types for mapping
type ProtocolsS3UserGetDataModelONTAP struct {
	SVM            svm     `mapstructure:"svm"`
	Name           string  `mapstructure:"name"`
	Comment        string  `mapstructure:"comment"`
	AccessKey      string  `mapstructure:"access_key"`
	SecretKey      string  `mapstructure:"secret_key,omitempty"`
	KeyExpiryTime  string  `mapstructure:"key_expiry_time,omitempty"`
	KeyTimeToLive  string  `mapstructure:"key_time_to_live,omitempty"`
	ID             int64   `mapstructure:"id"`
}

// ProtocolsS3UserResourceBodyDataModel describes the resource body data model using go types for mapping
type ProtocolsS3UserResourceBodyDataModel struct {
	Name            *string  `mapstructure:"name,omitempty"`
	Comment         *string  `mapstructure:"comment,omitempty"`
	AccessKey       *string  `mapstructure:"access_key,omitempty"`
	SecretKey       *string  `mapstructure:"secret_key,omitempty"`
	KeyTimeToLive   *string  `mapstructure:"key_time_to_live,omitempty"`
	RegenerateKeys  *bool    `mapstructure:"regenerate_keys,omitempty"`
	DeleteKeys 		*bool    `mapstructure:"delete_keys,omitempty"`
}

// S3UsersFilterModel describes the filter model for S3 users
type S3UsersFilterModel struct {
	SVMName  string  `mapstructure:"svm.name"`
	Name     string  `mapstructure:"name"`
}

// GetProtocolsS3User to get S3 user info given its name
func GetProtocolsS3User(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmUUID string, userName string, version versionModelONTAP) (*ProtocolsS3UserGetDataModelONTAP, error) {
	if version.Generation == 9 && version.Major < 7 {
		return nil, errorHandler.MakeAndReportError(
			"error reading S3 user info",
			"protocols/s3/services/{svm.uuid}/users API supported on ONTAP version 9.7 or later",
		)
	}
	api := fmt.Sprintf("protocols/s3/services/%s/users", svmUUID)

	query := r.NewQuery()
	query.Set("name", userName)
	var fields = []string{
		"svm",
		"name",
		"comment",
		"access_key",
	}
	if version.Generation == 9 && version.Major >= 14 {
		fields = append(fields, "key_expiry_time", "key_time_to_live")
	}
	query.Fields(fields)

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error reading S3 user info",
			fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP ProtocolsS3UserGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response),
		)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read s3_user data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetProtocolsS3Users to get a list of S3 users info for given filter options
func GetProtocolsS3Users(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmUUID string, userName string, version versionModelONTAP) ([]ProtocolsS3UserGetDataModelONTAP, error) {
	if version.Generation == 9 && version.Major < 7 {
		return nil, errorHandler.MakeAndReportError(
			"error reading S3 users info",
			"protocols/s3/services/{svm.uuid}/users API supported on ONTAP version 9.7 or later",
		)
	}
	
	api := fmt.Sprintf("protocols/s3/services/%s/users", svmUUID)
	query := r.NewQuery()
	if userName != "" {
		query.Set("name", userName)
	}
	var fields = []string{
		"svm",
		"name",
		"comment",
		"access_key",
	}
	if version.Generation == 9 && version.Major >= 14 {
		fields = append(fields, "key_expiry_time", "key_time_to_live")
	}
	query.Fields(fields)

	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error reading S3 users info",
			fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP []ProtocolsS3UserGetDataModelONTAP
	for _, info := range response {
		var record ProtocolsS3UserGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(
				fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info),
			)
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read s3_users data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateProtocolsS3User to create a S3 user
func CreateProtocolsS3User(errorHandler *utils.ErrorHandler, r restclient.RestClient, request ProtocolsS3UserResourceBodyDataModel, svmUUID string) (*ProtocolsS3UserGetDataModelONTAP, error) {
	api := fmt.Sprintf("protocols/s3/services/%s/users", svmUUID)
	
	var body map[string]interface{}
	if err := mapstructure.Decode(request, &body); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error encoding protocols/s3/services/{svm.uuid}/users body",
			fmt.Sprintf("error on encoding %s body: %s, request: %#v", api, err, request),
		)
	}
	
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, body)
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error creating S3 user",
			fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP ProtocolsS3UserGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error decoding S3 user info",
			fmt.Sprintf("error on decoding protocols/s3/services/{svm.uuid}/users info: %s, statusCode %d, response %#v", err, statusCode, response),
		)
	}

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Created s3_user resource: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteProtocolsS3User to delete a S3 user
func DeleteProtocolsS3User(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmUUID string, userName string) error {
	api := fmt.Sprintf("protocols/s3/services/%s/users/%s", svmUUID, userName)
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError(
			"error deleting S3 user",
			fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode),
		)
	}
	return nil
}

// UpdateProtocolsS3User to modify a S3 user
func UpdateProtocolsS3User(errorHandler *utils.ErrorHandler, r restclient.RestClient, request ProtocolsS3UserResourceBodyDataModel, svmUUID string, userName string) (*ProtocolsS3UserGetDataModelONTAP, error) {
	api := fmt.Sprintf("protocols/s3/services/%s/users/%s", svmUUID, userName)
	
	var body map[string]interface{}
	if err := mapstructure.Decode(request, &body); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error encoding protocols/s3/services/{svm.uuid}/users/{userName} body",
			fmt.Sprintf("error on encoding %s body: %s, request: %#v", api, err, request),
		)
	}
	
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallUpdateMethod(api, query, body)
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error modifying S3 user",
			fmt.Sprintf("error on PATCH %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	// action-style PATCH operations (e.g. delete_keys) can succeed
	// while returning no records, even if return_records=true
	if len(response.Records) == 0 {
		tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Updated s3_user resource: empty records response for PATCH %s, statusCode %d", api, statusCode))
		return &ProtocolsS3UserGetDataModelONTAP{}, nil
	}

	var dataONTAP ProtocolsS3UserGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error decoding S3 user info",
			fmt.Sprintf("error on decoding protocols/s3/services/{svm.uuid}/users/{user} info: %s, statusCode %d, response %#v", err, statusCode, response),
		)
	}

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Updated s3_user resource: %#v", dataONTAP))
	return &dataONTAP, nil
}
