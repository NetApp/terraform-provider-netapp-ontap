package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// SecurityLoginMessageGetDataModelONTAP describes the GET record data model using go types for mapping.
type SecurityLoginMessageGetDataModelONTAP struct {
	Message            string `mapstructure:"message"`
	SVM                svm    `mapstructure:"svm"`
	Scope              string `mapstructure:"scope"`
	Banner             string `mapstructure:"banner"`
	ShowClusterMessage bool   `mapstructure:"show_cluster_message"`
	UUID               string `mapstructure:"uuid"`
}

// SecurityLoginMessageResourceBodyDataModelONTAP describes the body data model using go types for mapping. Both svm and scope are not allowed to be updated.
type SecurityLoginMessageResourceBodyDataModelONTAP struct {
	Banner             string `mapstructure:"banner"`
	Message            string `mapstructure:"message"`
	ShowClusterMessage bool   `mapstructure:"show_cluster_message"`
}

// SecurityLoginMessageDataSourceFilterModel describes the data source data model for queries.
type SecurityLoginMessageDataSourceFilterModel struct {
	Banner  string `mapstructure:"banner"`
	Message string `mapstructure:"message"`
	Scope   string `mapstructure:"scope"`
	SVMName string `mapstructure:"svm.name"`
}

// GetSecurityLoginMessage to get security_login_message info
// Retrieves the login banner and messages of the day (MOTD) configured in the cluster and in a specific SVM.
func GetSecurityLoginMessage(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmName string) (*SecurityLoginMessageGetDataModelONTAP, error) {
	api := "security/login/messages"
	query := r.NewQuery()

	if svmName == "" {
		query.Set("scope", "cluster")
	} else {
		query.Set("svm.name", svmName)
		query.Set("scope", "svm")
	}
	query.Fields([]string{"show_cluster_message", "svm.name", "uuid", "scope", "banner", "message"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading security_login_message info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP SecurityLoginMessageGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read security_login_message data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetSecurityLoginMessages to get security_login_message info for all resources matching a filter
func GetSecurityLoginMessages(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *SecurityLoginMessageDataSourceFilterModel) ([]SecurityLoginMessageGetDataModelONTAP, error) {
	api := "security/login/messages"
	query := r.NewQuery()
	query.Fields([]string{"banner", "svm.name", "scope", "show_cluster_message", "uuid", "message"})
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding security_login_messages filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading security_login_messages info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []SecurityLoginMessageGetDataModelONTAP
	for _, info := range response {
		var record SecurityLoginMessageGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read security_login_messages data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// UpdateSecurityLoginMessage to update security_login_message
func UpdateSecurityLoginMessage(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string, body SecurityLoginMessageResourceBodyDataModelONTAP) error {
	api := "security/login/messages"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return errorHandler.MakeAndReportError("error encoding security_login_message body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Update security login message: %#v", body))
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, _, err := r.CallUpdateMethod(api+"/"+uuid, query, bodyMap)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating security_login_message", fmt.Sprintf("error on PUT %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
