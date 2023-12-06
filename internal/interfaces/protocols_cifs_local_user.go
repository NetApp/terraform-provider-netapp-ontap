package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// TODO:
// copy this file to match you data source (should match internal/interfaces/protocols_cifs_local_user.go)
// replace CifsLocalUser with the name of the resource, following go conventions, eg IPInterface
// replace protocols_cifs_local_user with the name of the resource, for logging purposes, eg ip_interface
// replace api_url with API, eg ip/interfaces
// delete these 5 lines

// CifsLocalUserGetDataModelONTAP describes the GET record data model using go types for mapping.
type CifsLocalUserGetDataModelONTAP struct {
	Name            string       `mapstructure:"name"`
	SID             string       `mapstructure:"sid"`
	SVM             svm          `mapstructure:"svm"`
	FullName        string       `mapstructure:"full_name"`
	Description     string       `mapstructure:"description"`
	Membership      []membership `mapstructure:"membership"`
	AccountDisabled bool         `mapstructure:"account_disabled"`
}

// Membership describes the membership data model using go types for mapping.
type membership struct {
	Name string `mapstructure:"name"`
}

// CifsLocalUserDataSourceFilterModel describes the data source data model for queries.
type CifsLocalUserDataSourceFilterModel struct {
	Name    string `mapstructure:"name"`
	SVMName string `mapstructure:"svm.name"`
}

// CifsLocalUserResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type CifsLocalUserResourceBodyDataModelONTAP struct {
	Name string `mapstructure:"name"`
	SVM  svm    `mapstructure:"svm"`
}

// GetCifsLocalUserByName to get protocols_cifs_local_user info
func GetCifsLocalUserByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmName string) (*CifsLocalUserGetDataModelONTAP, error) {
	api := "protocols/cifs/local-users"
	query := r.NewQuery()
	query.Set("name", name)
	query.Set("svm.name", svmName)

	query.Fields([]string{"name", "svm.name", "full_name", "description", "membership", "account_disabled"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_cifs_local_user info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP CifsLocalUserGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_cifs_local_user data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetCifsLocalUsers to get protocols_cifs_local_user info for all resources matching a filter
func GetCifsLocalUsers(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *CifsLocalUserDataSourceFilterModel) ([]CifsLocalUserGetDataModelONTAP, error) {
	api := "protocols/cifs/local-users"
	query := r.NewQuery()
	query.Fields([]string{"name", "svm.name", "full_name", "description", "membership", "account_disabled"})

	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding protocols_cifs_local_users filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_cifs_local_users info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []CifsLocalUserGetDataModelONTAP
	for _, info := range response {
		var record CifsLocalUserGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_cifs_local_users data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateCifsLocalUser to create protocols_cifs_local_user
func CreateCifsLocalUser(errorHandler *utils.ErrorHandler, r restclient.RestClient, body CifsLocalUserResourceBodyDataModelONTAP) (*CifsLocalUserGetDataModelONTAP, error) {
	api := "protocols/cifs/local-users"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding protocols_cifs_local_user body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating protocols_cifs_local_user", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP CifsLocalUserGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding protocols_cifs_local_user info", fmt.Sprintf("error on decode storage/protocols_cifs_local_users info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create protocols_cifs_local_user source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteCifsLocalUser to delete protocols_cifs_local_user
func DeleteCifsLocalUser(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := "protocols/cifs/local-users"
	statusCode, _, err := r.CallDeleteMethod(api+"/"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting protocols_cifs_local_user", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
