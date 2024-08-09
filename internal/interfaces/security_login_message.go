package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// TODO:
// copy this file to match you data source (should match internal/interfaces/security_login_message.go)
// replace SecurityLoginMessage with the name of the resource, following go conventions, eg IPInterface
// replace security_login_message with the name of the resource, for logging purposes, eg ip_interface
// replace api_url with API, eg ip/interfaces
// delete these 5 lines

// SecurityLoginMessageGetDataModelONTAP describes the GET record data model using go types for mapping.
type SecurityLoginMessageGetDataModelONTAP struct {
	Message            string `mapstructure:"message"`
	SVM                svm    `mapstructure:"svm"`
	Scope              string `mapstructure:"scope"`
	Banner             string `mapstructure:"banner"`
	ShowClusterMessage bool   `mapstructure:"show_cluster_message"`
	UUID               string `mapstructure:"uuid"`
}

// SecurityLoginMessageResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type SecurityLoginMessageResourceBodyDataModelONTAP struct {
	Name string `mapstructure:"name"`
	SVM  svm    `mapstructure:"svm"`
}

// SecurityLoginMessageDataSourceFilterModel describes the data source data model for queries.
type SecurityLoginMessageDataSourceFilterModel struct {
	Banner  string `mapstructure:"banner"`
	Message string `mapstructure:"message"`
	Scope   string `mapstructure:"scope"`
	SVMName string `mapstructure:"svm.name"`
}

// GetSecurityLoginMessageByName to get security_login_message info
func GetSecurityLoginMessageByBannerMotd(errorHandler *utils.ErrorHandler, r restclient.RestClient, banner string, message string, svmName string) (*SecurityLoginMessageGetDataModelONTAP, error) {
	api := "security/login/messages"
	query := r.NewQuery()
	if message != "" {
		query.Set("message", message)
	}
	if banner != "" {
		query.Set("banner", banner)
	}
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

// CreateSecurityLoginMessage to create security_login_message
func CreateSecurityLoginMessage(errorHandler *utils.ErrorHandler, r restclient.RestClient, body SecurityLoginMessageResourceBodyDataModelONTAP) (*SecurityLoginMessageGetDataModelONTAP, error) {
	api := "api_url"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding security_login_message body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating security_login_message", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP SecurityLoginMessageGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding security_login_message info", fmt.Sprintf("error on decode storage/security_login_messages info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create security_login_message source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteSecurityLoginMessage to delete security_login_message
func DeleteSecurityLoginMessage(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := "api_url"
	statusCode, _, err := r.CallDeleteMethod(api+"/"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting security_login_message", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
