package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// ProtocolsCIFSShareAclGetDataModelONTAP describes the GET record data model using go types for mapping.
type ProtocolsCIFSShareAclGetDataModelONTAP struct {
	Name        string `mapstructure:"name"`
	UUID        string `mapstructure:"uuid"`
	UserOrGroup string `mapstructure:"user_or_group"`
}

// ProtocolsCIFSShareAclResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type ProtocolsCIFSShareAclResourceBodyDataModelONTAP struct {
	// Name       string `mapstructure:"name"`
	// SVM        svm    `mapstructure:"svm"`
	Permission  string `mapstructure:"permission"`
	UserOrGroup string `mapstructure:"user_or_group"`
	Type        string `mapstructure:"type"`
}

// ProtocolsCIFSShareAclDataSourceFilterModel describes the data source data model for queries.
type ProtocolsCIFSShareAclDataSourceFilterModel struct {
	Name        string `mapstructure:"name"`
	SVMName     string `mapstructure:"svm.name"`
	UserOrGroup string `mapstructure:"user_or_group"`
}

// GetProtocolsCIFSShareAclByName to get protocols_cifs_share_acl info
func GetProtocolsCIFSShareAclByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmName string) (*ProtocolsCIFSShareAclGetDataModelONTAP, error) {
	api := "api_url"
	query := r.NewQuery()
	query.Set("name", name)
	if svmName == "" {
		query.Set("scope", "cluster")
	} else {
		query.Set("svm.name", svmName)
		query.Set("scope", "svm")
	}
	query.Fields([]string{"name", "svm.name", "ip", "scope"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_cifs_share_acl info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ProtocolsCIFSShareAclGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_cifs_share_acl data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetProtocolsCIFSShareAcls to get protocols_cifs_share_acl info for all resources matching a filter
func GetProtocolsCIFSShareAcls(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *ProtocolsCIFSShareAclDataSourceFilterModel, svmName string, shareName string) ([]ProtocolsCIFSShareAclGetDataModelONTAP, error) {
	api := fmt.Sprintf("/protocols/cifs/shares/%s/%s/acls", svmName, shareName)
	query := r.NewQuery()
	query.Fields([]string{"name", "svm.name", "scope"})
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding protocols_cifs_share_acls filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_cifs_share_acls info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []ProtocolsCIFSShareAclGetDataModelONTAP
	for _, info := range response {
		var record ProtocolsCIFSShareAclGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_cifs_share_acls data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateProtocolsCIFSShareAcl to create protocols_cifs_share_acl
func CreateProtocolsCIFSShareAcl(errorHandler *utils.ErrorHandler, r restclient.RestClient, body ProtocolsCIFSShareAclResourceBodyDataModelONTAP, svmID string, shareName string) (*ProtocolsCIFSShareAclGetDataModelONTAP, error) {
	api := fmt.Sprintf("/protocols/cifs/shares/%s/%s/acls", svmID, shareName)
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding protocols_cifs_share_acl body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating protocols_cifs_share_acl", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ProtocolsCIFSShareAclGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding protocols_cifs_share_acl info", fmt.Sprintf("error on decode storage/protocols_cifs_share_acls info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create protocols_cifs_share_acl source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// UpdateProtocolsCIFSShareAcl to update protocols_cifs_share_acl
func UpdateProtocolsCIFSShareAcl(errorHandler *utils.ErrorHandler, r restclient.RestClient, body ProtocolsCIFSShareAclResourceBodyDataModelONTAP, svmID string, shareName string, userOrGroup string, aclType string) error {
	api := fmt.Sprintf("/protocols/cifs/shares/%s/%s/acls/%s/%s", svmID, shareName, userOrGroup, aclType)
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return errorHandler.MakeAndReportError("error encoding protocols_cifs_share_acl body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	delete(bodyMap, "type")          // type is not returned in the response
	delete(bodyMap, "user_or_group") // user_or_group is not returned in the response
	statusCode, _, err := r.CallUpdateMethod(api, nil, bodyMap)
	if err != nil {
		return errorHandler.MakeAndReportError("error creating protocols_cifs_share_acl", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}

// DeleteProtocolsCIFSShareAcl to delete protocols_cifs_share_acl
func DeleteProtocolsCIFSShareAcl(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmID string, shareName string, userOrGroup string, aclType string) error {
	api := fmt.Sprintf("/protocols/cifs/shares/%s/%s/acls/%s/%s", svmID, shareName, userOrGroup, aclType)
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting protocols_cifs_share_acl", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
