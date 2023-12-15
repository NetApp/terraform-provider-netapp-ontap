package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// CifsLocalGroupGetDataModelONTAP describes the GET record data model using go types for mapping.
type CifsLocalGroupGetDataModelONTAP struct {
	Name        string   `mapstructure:"name"`
	SID         string   `mapstructure:"sid"`
	SVM         svm      `mapstructure:"svm"`
	Description string   `mapstructure:"description"`
	Members     []Member `mapstructure:"members,omitempty"`
}

// Member
type Member struct {
	Name string `mapstructure:"name"`
}

// CifsLocalGroupResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type CifsLocalGroupResourceBodyDataModelONTAP struct {
	Name        string `mapstructure:"name,omitempty"`
	SVM         svm    `mapstructure:"svm,omitempty"`
	Description string `mapstructure:"description,omitempty"`
}

// CifsLocalGroupDataSourceFilterModel describes the data source data model for queries.
type CifsLocalGroupDataSourceFilterModel struct {
	Name    string `mapstructure:"name"`
	SVMName string `mapstructure:"svm.name"`
}

// GetCifsLocalGroupByName to get protocols_cifs_local_group info
func GetCifsLocalGroupByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmName string) (*CifsLocalGroupGetDataModelONTAP, error) {
	api := "protocols/cifs/local-groups"
	query := r.NewQuery()
	query.Set("name", name)
	query.Set("svm.name", svmName)

	query.Fields([]string{"name", "svm.name", "description", "members"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_cifs_local_group info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP CifsLocalGroupGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_cifs_local_group data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetCifsLocalGroups to get protocols_cifs_local_group info for all resources matching a filter
func GetCifsLocalGroups(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *CifsLocalGroupDataSourceFilterModel) ([]CifsLocalGroupGetDataModelONTAP, error) {
	api := "protocols/cifs/local-groups"
	query := r.NewQuery()
	query.Fields([]string{"name", "svm.name", "description", "members"})
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding protocols_cifs_local_groups filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_cifs_local_groups info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []CifsLocalGroupGetDataModelONTAP
	for _, info := range response {
		var record CifsLocalGroupGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_cifs_local_groups data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// GetCifsLocalGroup to get protocols_cifs_local_group info
func GetCifsLocalGroup(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmid string, sid string) (*CifsLocalGroupGetDataModelONTAP, error) {
	api := "protocols/cifs/local-groups"
	query := r.NewQuery()

	query.Fields([]string{"name", "svm.name", "description", "members"})
	statusCode, response, err := r.GetNilOrOneRecord(api+"/"+svmid+"/"+sid, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_cifs_local_group info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP CifsLocalGroupGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_cifs_local_group data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// CreateCifsLocalGroup to create protocols_cifs_local_group
func CreateCifsLocalGroup(errorHandler *utils.ErrorHandler, r restclient.RestClient, body CifsLocalGroupResourceBodyDataModelONTAP) (*CifsLocalGroupGetDataModelONTAP, error) {
	api := "protocols/cifs/local-groups"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding protocols_cifs_local_group body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating protocols_cifs_local_group", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP CifsLocalGroupGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding protocols_cifs_local_group info", fmt.Sprintf("error on decode storage/protocols_cifs_local_groups info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create protocols_cifs_local_group source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteCifsLocalGroup to delete protocols_cifs_local_group
func DeleteCifsLocalGroup(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmid string, uuid string) error {
	api := "protocols/cifs/local-groups"
	statusCode, _, err := r.CallDeleteMethod(api+"/"+svmid+"/"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting protocols_cifs_local_group", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}

// UpdateCifsLocalGroup to update protocols_cifs_local_group
func UpdateCifsLocalGroup(errorHandler *utils.ErrorHandler, r restclient.RestClient, body CifsLocalGroupResourceBodyDataModelONTAP, svmid string, uuid string) (*CifsLocalGroupGetDataModelONTAP, error) {
	api := "protocols/cifs/local-groups"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding protocols_cifs_local_group body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallUpdateMethod(api+"/"+svmid+"/"+uuid, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error updating protocols_cifs_local_group", fmt.Sprintf("error on PUT %s: %s, statusCode %d", api, err, statusCode))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("$$$$$$$Update response status=%v, response=%v", statusCode, response))

	// var dataONTAP CifsLocalGroupGetDataModelONTAP
	// if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
	// 	return nil, errorHandler.MakeAndReportError("### error decoding protocols_cifs_local_group info", fmt.Sprintf("error on decode storage/protocols_cifs_local_groups info: %s, statusCode %d, response %#v", err, statusCode, response))
	// }
	// tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Update protocols_cifs_local_group source - udata: %#v", dataONTAP))
	return nil, nil
}
