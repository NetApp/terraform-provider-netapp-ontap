package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// NameServicesUnixGroupGetDataModelONTAP describes the GET record data model using go types for mapping
type NameServicesUnixGroupGetDataModelONTAP struct {
	SVM   svm                `mapstructure:"svm"`
	Name  string             `mapstructure:"name"`
	Users []UserGetDataModel `mapstructure:"users"`
	ID    int64              `mapstructure:"id"`
}

// UserGetDataModel describes the GET record data model using go types for mapping
// defined in protocols_s3_group.go
// type UserGetDataModel struct {
// 	Name string `mapstructure:"name"`
// }

// NameServicesUnixGroupResourceBodyDataModel describes the resource body data model using go types for mapping
type NameServicesUnixGroupResourceBodyDataModel struct {
	SVM  svm     `mapstructure:"svm,omitempty"`
	Name *string `mapstructure:"name,omitempty"`
	ID	 *int64  `mapstructure:"id,omitempty"`
}

// UnixGroupUsersResourceBodyDataModel describes the body for adding users to a unix group.
// expected payload shape: {"records": [{"name": "unix_user1"}, ...]}
type UnixGroupUsersResourceBodyDataModel struct {
	Records []UnixGroupUserRecord `mapstructure:"records,omitempty"`
}

// UnixGroupUserRecord describes a single user record for the unix group users endpoint.
// expected payload shape: {"records": [{"name": "unix_user1"}, ...]}
type UnixGroupUserRecord struct {
	Name string `mapstructure:"name" json:"name"`
}

// NameServicesUnixGroupsFilterModel describes the data source filter model for UNIX groups
type NameServicesUnixGroupsFilterModel struct {
	SVMName string `mapstructure:"svm.name"`
	Name    string `mapstructure:"name"`
}

// GetNameServicesUnixGroup to get UNIX group info given its name
func GetNameServicesUnixGroup(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmName string, groupName string, version versionModelONTAP) (*NameServicesUnixGroupGetDataModelONTAP, error) {
	if (version.Generation == 9 && version.Major < 9) || version.Generation > 9 {
		return nil, errorHandler.MakeAndReportError(
			"error reading UNIX group info",
			"name-services/unix-groups API supported on ONTAP version 9.9 or later",
		)
	}
	api := "name-services/unix-groups"

	query := r.NewQuery()
	query.Set("svm.name", svmName)
	query.Set("name", groupName)
	query.Fields([]string{
		"svm",
		"name",
		"id",
		"users.name",
	})

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error reading UNIX group info",
			fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP NameServicesUnixGroupGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response),
		)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read unix_group data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetNameServicesUnixGroups to get a list of UNIX groups info for given filter options
func GetNameServicesUnixGroups(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *NameServicesUnixGroupsFilterModel, version versionModelONTAP) ([]NameServicesUnixGroupGetDataModelONTAP, error) {
	if (version.Generation == 9 && version.Major < 9) || version.Generation > 9 {
		return nil, errorHandler.MakeAndReportError(
			"error reading UNIX groups info",
			"name-services/unix-groups API supported on ONTAP version 9.9 or later",
		)
	}
	
	api := "name-services/unix-groups"
	query := r.NewQuery()
	query.Fields([]string{
		"svm",
		"name",
		"id",
		"users.name",
	})

	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError(
				"error encoding UNIX groups filter info",
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
			"error reading UNIX groups info",
			fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP []NameServicesUnixGroupGetDataModelONTAP
	for _, info := range response {
		var record NameServicesUnixGroupGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(
				fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info),
			)
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read unix_groups data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateNameServicesUnixGroup to create a Unix group
func CreateNameServicesUnixGroup(errorHandler *utils.ErrorHandler, r restclient.RestClient, request NameServicesUnixGroupResourceBodyDataModel) (*NameServicesUnixGroupGetDataModelONTAP, error) {
	api := "name-services/unix-groups"
	
	var body map[string]interface{}
	if err := mapstructure.Decode(request, &body); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error encoding name-services/unix-groups body",
			fmt.Sprintf("error on encoding %s body: %s, request: %#v", api, err, request),
		)
	}
	
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, body)
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error creating UNIX group",
			fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP NameServicesUnixGroupGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error decoding UNIX group info",
			fmt.Sprintf("error on decoding name-services/unix-groups info: %s, statusCode %d, response %#v", err, statusCode, response),
		)
	}

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Created UNIX group resource: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteNameServicesUnixGroup to delete a UNIX group
func DeleteNameServicesUnixGroup(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmUUID string, groupName string) error {
	api := fmt.Sprintf("name-services/unix-groups/%s/%s", svmUUID, groupName)
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError(
			"error deleting UNIX group",
			fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode),
		)
	}
	return nil
}

// UpdateNameServicesUnixGroup to modify a UNIX group
func UpdateNameServicesUnixGroup(errorHandler *utils.ErrorHandler, r restclient.RestClient, request NameServicesUnixGroupResourceBodyDataModel, svmUUID string, groupName string) error {
	api := fmt.Sprintf("name-services/unix-groups/%s/%s", svmUUID, groupName)
	
	var body map[string]interface{}
	if err := mapstructure.Decode(request, &body); err != nil {
		return errorHandler.MakeAndReportError(
			"error encoding name-services/unix-groups body",
			fmt.Sprintf("error on encoding %s body: %s, request: %#v", api, err, request),
		)
	}
	
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, _, err := r.CallUpdateMethod(api, query, body)
	if err != nil {
		return errorHandler.MakeAndReportError(
			"error modifying UNIX group",
			fmt.Sprintf("error on PATCH %s: %s, statusCode %d", api, err, statusCode),
		)
	}
	return nil
}

// helper function AddUnixGroupUsers to add users to a Unix group
func AddUnixGroupUsers(errorHandler *utils.ErrorHandler, r restclient.RestClient, request UnixGroupUsersResourceBodyDataModel, svmUUID string, groupName string) error {
	api := fmt.Sprintf("name-services/unix-groups/%s/%s/users", svmUUID, groupName)
	
	var body map[string]interface{}
	if err := mapstructure.Decode(request, &body); err != nil {
		return errorHandler.MakeAndReportError(
			"error encoding name-services/unix-groups body",
			fmt.Sprintf("error on encoding %s body: %s, request: %#v", api, err, request),
		)
	}
	
	statusCode, _, err := r.CallCreateMethod(api, nil, body)
	if err != nil {
		return errorHandler.MakeAndReportError(
			"error adding users to UNIX group",
			fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode),
		)
	}
	return nil
}

// helper function DeleteUnixGroupUser to delete a UNIX group user
func DeleteUnixGroupUser(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmUUID string, groupName string, userName string) error {
	api := fmt.Sprintf("name-services/unix-groups/%s/%s/users/%s", svmUUID, groupName, userName)
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError(
			fmt.Sprintf("error deleting UNIX group user %s", userName),
			fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode),
		)
	}
	return nil
}
