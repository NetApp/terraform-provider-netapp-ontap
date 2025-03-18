package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// ProtocolsIscsiServiceGetDataModelONTAP describes the GET record data model using go types for mapping.
// https://docs.netapp.com/us-en/ontap-restapi/ontap/get-protocols-san-iscsi-services-.html#response
type ProtocolsIscsiServiceGetDataModelONTAP struct {
	Enabled bool                        `mapstructure:"enabled"`
	SVM     SvmDataModelONTAP           `mapstructure:"svm"`
	Target  ProtocolsIscsiServiceTarget `mapstructure:"target"`
}

// ProtocolsIscsiServiceResourceDataModelONTAP describes the body data model using go types for mapping.
// https://docs.netapp.com/us-en/ontap-restapi/ontap/post-protocols-san-iscsi-services.html#request-body
type ProtocolsIscsiServiceResourceDataModelONTAP struct {
	Enabled bool                        `mapstructure:"enabled,omitempty"`
	SVM     SvmDataModelONTAP           `mapstructure:"svm"`
	Target  ProtocolsIscsiServiceTarget `mapstructure:"target"`
}

// ProtocolsIscsiServiceTarget describes a target specifically for protocol iSCSI services.
type ProtocolsIscsiServiceTarget struct {
	Alias string `mapstructure:"alias,omitempty"`
	Name  string `mapstructure:"name,omitempty"`
}

// IscsiServicesFilterModel describes filter model
type IscsiServicesFilterModel struct {
	SVMName string `mapstructure:"svm.name"`
}

// Retrieve an iSCSI service
// https://docs.netapp.com/us-en/ontap-restapi/ontap/get-protocols-san-iscsi-services-.html
func GetProtocolsIscsiService(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmName string) (*ProtocolsIscsiServiceGetDataModelONTAP, error) {
	api := "protocols/san/iscsi/services"
	query := r.NewQuery()
	query.Set("svm.name", svmName)
	query.Fields([]string{
		"svm.uuid",
		"enabled",
		"target",
	})

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no iscsi service for svm '%s' found", svmName)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"Error Reading iSCSI Service Info",
			fmt.Sprintf("Error on GET %s: %s, statusCode %d.", api, err, statusCode),
		)
	}

	var dataONTAP ProtocolsIscsiServiceGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			fmt.Sprintf("Failed To Decode Response From GET %s", api),
			fmt.Sprintf("Error: %s, statusCode %d, response %#v.", err, statusCode, response),
		)
	}

	tflog.Debug(
		errorHandler.Ctx,
		fmt.Sprintf("Read protcols iscsi service data source: %#v", dataONTAP),
	)

	return &dataONTAP, nil
}

// Retrieve iSCSI services
// https://docs.netapp.com/us-en/ontap-restapi/ontap/get-protocols-san-iscsi-services.html
func GetProtocolsIscsiServices(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *IscsiServicesFilterModel) ([]ProtocolsIscsiServiceGetDataModelONTAP, error) {
	api := "protocols/san/iscsi/services"
	query := r.NewQuery()
	query.Fields([]string{
		"svm",
		"enabled",
		"target",
	})

	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError(
				"Error Encoding iSCSI Services Filter",
				fmt.Sprintf("Error on filter %#v: %s", filter, err),
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
			"Error Reading iSCSI Services Info",
			fmt.Sprintf("Error on GET %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP []ProtocolsIscsiServiceGetDataModelONTAP
	for _, info := range response {
		var record ProtocolsIscsiServiceGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(
				fmt.Sprintf("Failed To Decode Response From GET %s", api),
				fmt.Sprintf("Error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}

	tflog.Debug(
		errorHandler.Ctx,
		fmt.Sprintf("Read protcols iscsi services data source: %#v", dataONTAP),
	)

	return dataONTAP, nil
}

// Create an iSCSI service
// https://docs.netapp.com/us-en/ontap-restapi/ontap/post-protocols-san-iscsi-services.html
func CreateProtocolsIscsiService(errorHandler *utils.ErrorHandler, r restclient.RestClient, data ProtocolsIscsiServiceResourceDataModelONTAP) (*ProtocolsIscsiServiceGetDataModelONTAP, error) {
	api := "protocols/san/iscsi/services"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(data, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"Error Encoding iSCSI Service Body",
			fmt.Sprintf("Error on encoding %s body: %s, body: %#v", api, err, data),
		)
	}

	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"Error Creating iSCSI Services",
			fmt.Sprintf("Error on POST %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP ProtocolsIscsiServiceGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"Error Decoding iSCSI Services Info",
			fmt.Sprintf("Error on decode %s info: %s, statusCode %d, response %#v", api, err, statusCode, response),
		)
	}

	tflog.Debug(
		errorHandler.Ctx,
		fmt.Sprintf("Create protcols iscsi services: %#v", dataONTAP),
	)

	return &dataONTAP, nil
}

// Delete an iSCSI service
// https://docs.netapp.com/us-en/ontap-restapi/ontap/delete-protocols-san-iscsi-services-.html
func DeleteProtocolsIscsiService(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := "protocols/san/iscsi/services"
	statusCode, _, err := r.CallDeleteMethod(api+"/"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError(
			"Error Deleting iSCSI Service",
			fmt.Sprintf("Error on DELETE %s: %s, statusCode %d", api, err, statusCode),
		)
	}
	return nil
}

// Update an iSCSI service
// https://docs.netapp.com/us-en/ontap-restapi/ontap/patch-protocols-san-iscsi-services-.html
func UpdateProtocolsIscsiService(errorHandler *utils.ErrorHandler, r restclient.RestClient, request ProtocolsIscsiServiceResourceDataModelONTAP, uuid string) error {
	var body map[string]interface{}
	if err := mapstructure.Decode(request, &body); err != nil {
		return errorHandler.MakeAndReportError("error encoding NFS Services body", fmt.Sprintf("error on encoding NFS Services body: %s, body: %#v", err, request))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, _, err := r.CallUpdateMethod("protocols/san/iscsi/services/"+uuid, query, body)
	if err != nil {
		return errorHandler.MakeAndReportError("error modifying NFS Service", fmt.Sprintf("error on PATCH rotocols/nfs/services/s: %s, statusCode %d", err, statusCode))
	}
	return nil
}
