package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// SecurityRoleGetDataModelONTAP describes the GET record data model using go types for mapping.
type SecurityRoleGetDataModelONTAP struct {
	Name       string                   `mapstructure:"name"`
	UUID       string                   `mapstructure:"uuid"`
	Owner      SecurityRoleOwner        `mapstructure:"owner"`
	Privileges []SecurityRolePrivileges `mapstructure:"privileges"`
	Scope      string                   `mapstructure:"scope"`
	Builtin    bool                     `mapstructure:"builtin"`
}

type SecurityRolePrivileges struct {
	Access string `mapstructure:"access"`
	Path   string `mapstructure:"path"`
}

type SecurityRoleOwner struct {
	Name string `mapstructure:"name"`
	Id   string `mapstructure:"uuid"`
}

// SecurityRoleResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type SecurityRoleResourceBodyDataModelONTAP struct {
	Name string `mapstructure:"name"`
	SVM  svm    `mapstructure:"svm"`
}

// SecurityRoleDataSourceFilterModel describes the data source data model for queries.
type SecurityRoleDataSourceFilterModel struct {
	Name    string `mapstructure:"name"`
	SVMName string `mapstructure:"owner.name"`
	Scope   string `mapstructure:"scope"`
}

// GetSecurityRoleByName to get security_role info
func GetSecurityRoleByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmUUID string) (*SecurityRoleGetDataModelONTAP, error) {
	api := "security/roles/" + svmUUID + "/" + name
	query := r.NewQuery()
	query.Set("name", name)
	query.Fields([]string{"name", "scope", "owner", "privileges", "builtin"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading security_role info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP SecurityRoleGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read security_role data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetSecurityRoles to get security_role info for all resources matching a filter
func GetSecurityRoles(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *SecurityRoleDataSourceFilterModel) ([]SecurityRoleGetDataModelONTAP, error) {
	api := "security/roles"
	query := r.NewQuery()
	query.Fields([]string{"name", "scope", "owner", "privileges", "builtin"})
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding security_roles filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading security_roles info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []SecurityRoleGetDataModelONTAP
	for _, info := range response {
		var record SecurityRoleGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read security_roles data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateSecurityRole to create security_role
func CreateSecurityRole(errorHandler *utils.ErrorHandler, r restclient.RestClient, body SecurityRoleResourceBodyDataModelONTAP) (*SecurityRoleGetDataModelONTAP, error) {
	api := "api_url"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding security_role body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating security_role", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP SecurityRoleGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding security_role info", fmt.Sprintf("error on decode storage/security_roles info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create security_role source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteSecurityRole to delete security_role
func DeleteSecurityRole(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := "api_url"
	statusCode, _, err := r.CallDeleteMethod(api+"/"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting security_role", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
