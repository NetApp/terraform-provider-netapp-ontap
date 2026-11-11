package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// NameServicesUnixUserGetDataModelONTAP describes the GET record data model using go types for mapping
type NameServicesUnixUserGetDataModelONTAP struct {
	SVM        svm    `mapstructure:"svm"`
	Name       string `mapstructure:"name"`
	FullName   string `mapstructure:"full_name"`
	PrimaryGID int64  `mapstructure:"primary_gid"`
	ID         int64  `mapstructure:"id"`
}

// NameServicesUnixUserResourceBodyDataModel describes the resource body data model using go types for mapping
type NameServicesUnixUserResourceBodyDataModel struct {
	SVM        svm     `mapstructure:"svm,omitempty"`
	Name       *string `mapstructure:"name,omitempty"`
	FullName   *string `mapstructure:"full_name,omitempty"`
	PrimaryGID *int64  `mapstructure:"primary_gid,omitempty"`
	ID		   *int64  `mapstructure:"id,omitempty"`
}

// NameServicesUnixUsersFilterModel describes the data source filter model for UNIX users
type NameServicesUnixUsersFilterModel struct {
	SVMName string `mapstructure:"svm.name"`
	Name    string `mapstructure:"name"`
}

// GetNameServicesUnixUser to get UNIX user info given its name
func GetNameServicesUnixUser(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmName string, userName string, version versionModelONTAP) (*NameServicesUnixUserGetDataModelONTAP, error) {
	if (version.Generation == 9 && version.Major < 9) || version.Generation > 9 {
		return nil, errorHandler.MakeAndReportError(
			"error reading UNIX user info",
			"name-services/unix-users API supported on ONTAP version 9.9 or later",
		)
	}
	api := "name-services/unix-users"

	query := r.NewQuery()
	query.Set("svm.name", svmName)
	query.Set("name", userName)
	query.Fields([]string{
		"svm",
		"name",
		"full_name",
		"primary_gid",
		"id",
	})

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error reading UNIX user info",
			fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP NameServicesUnixUserGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response),
		)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read unix_user data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetNameServicesUnixUsers to get a list of UNIX users info for given filter options
func GetNameServicesUnixUsers(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *NameServicesUnixUsersFilterModel, version versionModelONTAP) ([]NameServicesUnixUserGetDataModelONTAP, error) {
	if (version.Generation == 9 && version.Major < 9) || version.Generation > 9 {
		return nil, errorHandler.MakeAndReportError(
			"error reading UNIX users info",
			"name-services/unix-users API supported on ONTAP version 9.9 or later",
		)
	}
	
	api := "name-services/unix-users"
	query := r.NewQuery()
	query.Fields([]string{
		"svm",
		"name",
		"full_name",
		"primary_gid",
		"id",
	})

	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError(
				"error encoding UNIX users filter info",
				fmt.Sprintf("error on filter %#v: %s", filter, err),
			)
		}
		query.SetValues(filterMap)
	}

	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error reading UNIX users info",
			fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP []NameServicesUnixUserGetDataModelONTAP
	for _, info := range response {
		var record NameServicesUnixUserGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(
				fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info),
			)
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read unix_users data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateNameServicesUnixUser to create a Unix user
func CreateNameServicesUnixUser(errorHandler *utils.ErrorHandler, r restclient.RestClient, request NameServicesUnixUserResourceBodyDataModel) (*NameServicesUnixUserGetDataModelONTAP, error) {
	api := "name-services/unix-users"
	
	var body map[string]interface{}
	if err := mapstructure.Decode(request, &body); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error encoding name-services/unix-users body",
			fmt.Sprintf("error on encoding %s body: %s, request: %#v", api, err, request),
		)
	}
	
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, body)
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error creating UNIX user",
			fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP NameServicesUnixUserGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error decoding UNIX user info",
			fmt.Sprintf("error on decoding name-services/unix-users info: %s, statusCode %d, response %#v", err, statusCode, response),
		)
	}

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Created UNIX user resource: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteNameServicesUnixUser to delete a UNIX user
func DeleteNameServicesUnixUser(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmUUID string, userName string) error {
	api := fmt.Sprintf("name-services/unix-users/%s/%s", svmUUID, userName)
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError(
			"error deleting UNIX user",
			fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode),
		)
	}
	return nil
}

// UpdateNameServicesUnixUser to modify a UNIX user
func UpdateNameServicesUnixUser(errorHandler *utils.ErrorHandler, r restclient.RestClient, request NameServicesUnixUserResourceBodyDataModel, svmUUID string, userName string) error {
	api := fmt.Sprintf("name-services/unix-users/%s/%s", svmUUID, userName)
	
	var body map[string]interface{}
	if err := mapstructure.Decode(request, &body); err != nil {
		return errorHandler.MakeAndReportError(
			"error encoding name-services/unix-users body",
			fmt.Sprintf("error on encoding %s body: %s, request: %#v", api, err, request),
		)
	}
	
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, _, err := r.CallUpdateMethod(api, query, body)
	if err != nil {
		return errorHandler.MakeAndReportError(
			"error modifying UNIX user",
			fmt.Sprintf("error on PATCH %s: %s, statusCode %d", api, err, statusCode),
		)
	}
	return nil
}
