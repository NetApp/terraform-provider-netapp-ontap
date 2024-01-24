package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// CifsUserGroupPrivilegeGetDataModelONTAP describes the GET record data model using go types for mapping.
type CifsUserGroupPrivilegeGetDataModelONTAP struct {
	Name string `mapstructure:"name"`
	// UUID string `mapstructure:"uuid"`
	SVM        svm      `mapstructure:"svm"`
	Privileges []string `mapstructure:"privileges"`
}

// CifsUserGroupPrivilegeResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type CifsUserGroupPrivilegeResourceBodyDataModelONTAP struct {
	Name string `mapstructure:"name"`
	SVM  svm    `mapstructure:"svm"`
}

// CifsUserGroupPrivilegeDataSourceFilterModel describes the data source data model for queries.
type CifsUserGroupPrivilegeDataSourceFilterModel struct {
	Name       string `mapstructure:"name"`
	SVMName    string `mapstructure:"svm.name"`
	Privileges string `mapstructure:"privileges"` //only support one privilege search
}

// GetCifsUserGroupPrivilegeByName to get protocols_cifs_user_group_privilege info
func GetCifsUserGroupPrivilegeByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmName string) (*CifsUserGroupPrivilegeGetDataModelONTAP, error) {
	api := "protocols/cifs/users-and-groups/privileges"
	query := r.NewQuery()
	query.Set("name", name)
	query.Set("svm.name", svmName)

	query.Fields([]string{"name", "svm.name", "privileges"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_cifs_user_group_privilege info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP CifsUserGroupPrivilegeGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_cifs_user_group_privilege data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetCifsUserGroupPrivileges to get protocols_cifs_user_group_privilege info for all resources matching a filter
func GetCifsUserGroupPrivileges(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *CifsUserGroupPrivilegeDataSourceFilterModel) ([]CifsUserGroupPrivilegeGetDataModelONTAP, error) {
	api := "protocols/cifs/users-and-groups/privileges"
	query := r.NewQuery()
	query.Fields([]string{"name", "svm.name", "privileges"})
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding protocols_cifs_user_group_privileges filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_cifs_user_group_privileges info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []CifsUserGroupPrivilegeGetDataModelONTAP
	for _, info := range response {
		var record CifsUserGroupPrivilegeGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_cifs_user_group_privileges data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateCifsUserGroupPrivilege to create protocols_cifs_user_group_privilege
func CreateCifsUserGroupPrivilege(errorHandler *utils.ErrorHandler, r restclient.RestClient, body CifsUserGroupPrivilegeResourceBodyDataModelONTAP) (*CifsUserGroupPrivilegeGetDataModelONTAP, error) {
	api := "protocols/cifs/users-and-groups/privileges"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding protocols_cifs_user_group_privilege body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating protocols_cifs_user_group_privilege", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP CifsUserGroupPrivilegeGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding protocols_cifs_user_group_privilege info", fmt.Sprintf("error on decode storage/protocols_cifs_user_group_privileges info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create protocols_cifs_user_group_privilege source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteCifsUserGroupPrivilege to delete protocols_cifs_user_group_privilege
func DeleteCifsUserGroupPrivilege(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := "protocols/cifs/users-and-groups/privileges"
	statusCode, _, err := r.CallDeleteMethod(api+"/"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting protocols_cifs_user_group_privilege", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
