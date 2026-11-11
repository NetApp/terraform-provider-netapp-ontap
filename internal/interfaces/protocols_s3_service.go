package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// ProtocolsS3ServiceGetDataModelONTAP describes the GET record data model using go types for mapping
type ProtocolsS3ServiceGetDataModelONTAP struct {
	SVM             svm                  `mapstructure:"svm"`
	Name            string               `mapstructure:"name"`
	Enabled         bool                 `mapstructure:"enabled"`
	Comment         string               `mapstructure:"comment"`
	Certificate     CertificateDataModel `mapstructure:"certificate"`
	IsHTTPEnabled   bool                 `mapstructure:"is_http_enabled"`
	IsHTTPSEnabled  bool                 `mapstructure:"is_https_enabled"`
	Port 			int64                `mapstructure:"port"`
	SecurePort 	    int64                `mapstructure:"secure_port"`
	ID              string               `mapstructure:"ID"`
}

// CertificateDataModel describes the resource data model.
type CertificateDataModel struct {
	Name string `mapstructure:"name,omitempty"`
}

// ProtocolsS3ServiceResourceBodyDataModel describes the resource body data model using go types for mapping
type ProtocolsS3ServiceResourceBodyDataModel struct {
	SVM             svm                   `mapstructure:"svm,omitempty"`
	Name            string                `mapstructure:"name"`
	Enabled         *bool                 `mapstructure:"enabled,omitempty"`
	Comment         *string               `mapstructure:"comment,omitempty"`
	Certificate     *CertificateDataModel `mapstructure:"certificate,omitempty"`
	IsHTTPEnabled   *bool                 `mapstructure:"is_http_enabled,omitempty"`
	IsHTTPSEnabled  *bool                 `mapstructure:"is_https_enabled,omitempty"`
	Port            *int64                `mapstructure:"port,omitempty"`
	SecurePort      *int64                `mapstructure:"secure_port,omitempty"`
}

// ProtocolsS3ServiceResourceBodyDataModel describes the data source filter model for protocols S3 service
type ProtocolsS3ServicesFilterModel struct {
	SVMName string `mapstructure:"svm.name"`
	Name    string `mapstructure:"name"`
}

// GetS3Server to get S3 server configuration for given SVM
func GetS3Server(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmName string, version versionModelONTAP) (*ProtocolsS3ServiceGetDataModelONTAP, error) {
	if version.Generation == 9 && version.Major < 8 {
		return nil, errorHandler.MakeAndReportError(
			"error reading S3 server info",
			"protocols/s3/services API supported on ONTAP version 9.8 or later",
		)
	}
	api := "protocols/s3/services"

	query := r.NewQuery()
	query.Set("svm.name", svmName)
	var fields = []string{
		"svm",
		"name",
		"enabled",
		"comment",
		"certificate.name",
		"is_http_enabled",
		"is_https_enabled",
		"port",
		"secure_port",
	}
	query.Fields(fields)

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error reading S3 server info",
			fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP ProtocolsS3ServiceGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response),
		)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read s3_service data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetS3Servers to get S3 server configurations for given filter options
func GetS3Servers(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *ProtocolsS3ServicesFilterModel, version versionModelONTAP) ([]ProtocolsS3ServiceGetDataModelONTAP, error) {
	if version.Generation == 9 && version.Major < 8 {
		return nil, errorHandler.MakeAndReportError(
			"error reading S3 servers info",
			"protocols/s3/services API supported on ONTAP version 9.8 or later",
		)
	}
	api := "protocols/s3/services"

	query := r.NewQuery()
	var fields = []string{
		"svm",
		"name",
		"enabled",
		"comment",
		"certificate.name",
		"is_http_enabled",
		"is_https_enabled",
		"port",
		"secure_port",
	}
	query.Fields(fields)

	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError(
				"error encoding S3 servers filter info",
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
			"error reading S3 servers info",
			fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP []ProtocolsS3ServiceGetDataModelONTAP
	for _, info := range response {
		var record ProtocolsS3ServiceGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(
				fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info),
			)
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read s3_services data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateS3Server creates S3 server for a SVM
func CreateS3Server(errorHandler *utils.ErrorHandler, r restclient.RestClient, request ProtocolsS3ServiceResourceBodyDataModel) (*ProtocolsS3ServiceGetDataModelONTAP, error) {
	api := "protocols/s3/services"
	
	var body map[string]interface{}
	if err := mapstructure.Decode(request, &body); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error encoding protocols/s3/services body",
			fmt.Sprintf("error on encoding %s body: %s, request: %#v", api, err, request),
		)
	}
	
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, body)
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error creating S3 server",
			fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP ProtocolsS3ServiceGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error decoding S3 server response",
			fmt.Sprintf("error on decoding protocols/s3/services info: %s, statusCode %d, response %#v", err, statusCode, response),
		)
	}

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Created S3 server resource: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteS3Server deletes S3 server for a SVM
func DeleteS3Server(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmUUID string) error {
	api := fmt.Sprintf("protocols/s3/services/%s", svmUUID)
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError(
			"error deleting S3 server",
			fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode),
		)
	}
	return nil
}

// UpdateS3Server modifies S3 server for a SVM
func UpdateS3Server(errorHandler *utils.ErrorHandler, r restclient.RestClient, request ProtocolsS3ServiceResourceBodyDataModel, svmUUID string) error {
	api := fmt.Sprintf("protocols/s3/services/%s", svmUUID)
	
	var body map[string]interface{}
	if err := mapstructure.Decode(request, &body); err != nil {
		return errorHandler.MakeAndReportError(
			"error encoding protocols/s3/services/{svm.uuid} body",
			fmt.Sprintf("error on encoding %s body: %s, request: %#v", api, err, request),
		)
	}
	
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, _, err := r.CallUpdateMethod(api, query, body)
	if err != nil {
		return errorHandler.MakeAndReportError(
			"error modifying S3 server",
			fmt.Sprintf("error on PATCH %s: %s, statusCode %d", api, err, statusCode),
		)
	}
	return nil
}
