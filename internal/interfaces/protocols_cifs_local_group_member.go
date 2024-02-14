package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// CifsLocalGroupMemberGetDataModelONTAP describes the GET record data model using go types for mapping.
type CifsLocalGroupMemberGetDataModelONTAP struct {
	Name string `mapstructure:"name"`
	SVM  svm    `mapstructure:"svm"`
}

// CifsLocalGroupMemberResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type CifsLocalGroupMemberResourceBodyDataModelONTAP struct {
	Name string `mapstructure:"name"`
}

// CifsLocalGroupMemberDataSourceFilterModel describes the data source data model for queries.
type CifsLocalGroupMemberDataSourceFilterModel struct {
	Name    string `mapstructure:"name"` // Name of the local group name
	SVMName string `mapstructure:"svm.name"`
}

// GetCifsLocalGroupMemberByName to get protocols_cifs_local_group_member info
func GetCifsLocalGroupMemberByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmid string, groupid string, user string) (*CifsLocalGroupMemberGetDataModelONTAP, error) {
	api := "protocols/cifs/local-groups/" + svmid + "/" + groupid + "/members"
	query := r.NewQuery()
	query.Set("name", user)

	query.Fields([]string{"name", "svm.name"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_cifs_local_group_member info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP CifsLocalGroupMemberGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_cifs_local_group_member data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetCifsLocalGroupMembers to get protocols_cifs_local_group_members info for all members of a group under a svm
func GetCifsLocalGroupMembers(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmid string, groupid string) ([]CifsLocalGroupMemberGetDataModelONTAP, error) {
	api := "protocols/cifs/local-groups/" + svmid + "/" + groupid + "/members"
	query := r.NewQuery()
	query.Fields([]string{"name", "svm.name"})
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_cifs_local_group_members info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []CifsLocalGroupMemberGetDataModelONTAP
	for _, info := range response {
		var record CifsLocalGroupMemberGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_cifs_local_group_members data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateCifsLocalGroupMember to create protocols_cifs_local_group_member
func CreateCifsLocalGroupMember(errorHandler *utils.ErrorHandler, r restclient.RestClient, body CifsLocalGroupMemberResourceBodyDataModelONTAP, svmid string, groupid string) (*CifsLocalGroupMemberGetDataModelONTAP, error) {
	api := "protocols/cifs/local-groups/" + svmid + "/" + groupid + "/members"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding protocols_cifs_local_group_member body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating protocols_cifs_local_group_member", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP CifsLocalGroupMemberGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding protocols_cifs_local_group_member info", fmt.Sprintf("error on decode storage/protocols_cifs_local_group_members info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create protocols_cifs_local_group_member source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteCifsLocalGroupMember to delete protocols_cifs_local_group_member
func DeleteCifsLocalGroupMember(errorHandler *utils.ErrorHandler, r restclient.RestClient, body CifsLocalGroupMemberResourceBodyDataModelONTAP, svmid string, groupid string) error {
	api := "protocols/cifs/local-groups/" + svmid + "/" + groupid + "/members"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return errorHandler.MakeAndReportError("error encoding protocols_cifs_local_group_member body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	statusCode, _, err := r.CallDeleteMethod(api, nil, bodyMap)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting protocols_cifs_local_group_member", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
