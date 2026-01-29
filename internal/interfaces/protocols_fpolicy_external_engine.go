package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// ProtocolsFpolicyExternalEngineGetDataModelONTAP describes the GET record data model using go types for mapping.
type ProtocolsFpolicyExternalEngineGetDataModelONTAP struct {
	Name                      string                                `mapstructure:"name"`
	SVM                       svm                                   `mapstructure:"svm"`
	KeepAliveInterval         string                                `mapstructure:"keep_alive_interval,omitempty"`
	RequestCancelTimeout      string                                `mapstructure:"request_cancel_timeout,omitempty"`
	Certificate               ProtocolsFpolicyExternalEngineCertificate `mapstructure:"certificate,omitempty"`
	SessionTimeout            string                                `mapstructure:"session_timeout,omitempty"`
	RequestAbortTimeout       string                                `mapstructure:"request_abort_timeout,omitempty"`
	StatusRequestInterval     string                                `mapstructure:"status_request_interval,omitempty"`
	SSLOption                 string                                `mapstructure:"ssl_option,omitempty"`
	PrimaryServers            []string                              `mapstructure:"primary_servers,omitempty"`
	BufferSize                ProtocolsFpolicyExternalEngineBufferSize `mapstructure:"buffer_size,omitempty"`
	SecondaryServers          []string                              `mapstructure:"secondary_servers,omitempty"`
	Port                      int64                                 `mapstructure:"port,omitempty"`
	ServerProgressTimeout     string                                `mapstructure:"server_progress_timeout,omitempty"`
	Format                    string                                `mapstructure:"format,omitempty"`
	Type                      string                                `mapstructure:"type,omitempty"`
	Resiliency                ProtocolsFpolicyExternalEngineResiliency `mapstructure:"resiliency,omitempty"`
	MaxServerRequests         int64                                 `mapstructure:"max_server_requests,omitempty"`
	MaxConnectionRetries      int64                                 `mapstructure:"max_connection_retries,omitempty"`
}

// ProtocolsFpolicyExternalEngineCertificate describes the certificate data model
type ProtocolsFpolicyExternalEngineCertificate struct {
	SerialNumber string `mapstructure:"serial_number,omitempty"`
	Name         string `mapstructure:"name,omitempty"`
	Ca           string `mapstructure:"ca,omitempty"`
}

// ProtocolsFpolicyExternalEngineBufferSize describes the buffer size data model
type ProtocolsFpolicyExternalEngineBufferSize struct {
	SendBuffer int `mapstructure:"send_buffer,omitempty"`
	RecvBuffer int `mapstructure:"recv_buffer,omitempty"`
}

// ProtocolsFpolicyExternalEngineResiliency describes the resiliency data model
type ProtocolsFpolicyExternalEngineResiliency struct {
	DirectoryPath     string `mapstructure:"directory_path,omitempty"`
	RetentionDuration string `mapstructure:"retention_duration,omitempty"`
	Enabled           bool   `mapstructure:"enabled"`
}

// ProtocolsFpolicyExternalEngineResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type ProtocolsFpolicyExternalEngineResourceBodyDataModelONTAP struct {
	Name                      string                                `mapstructure:"name,omitempty"`
	SVM                       svm                                   `mapstructure:"svm"`
	KeepAliveInterval         string                                `mapstructure:"keep_alive_interval,omitempty"`
	RequestCancelTimeout      string                                `mapstructure:"request_cancel_timeout,omitempty"`
	Certificate               ProtocolsFpolicyExternalEngineCertificate `mapstructure:"certificate,omitempty"`
	SessionTimeout            string                                `mapstructure:"session_timeout,omitempty"`
	RequestAbortTimeout       string                                `mapstructure:"request_abort_timeout,omitempty"`
	StatusRequestInterval     string                                `mapstructure:"status_request_interval,omitempty"`
	SSLOption                 string                                `mapstructure:"ssl_option,omitempty"`
	PrimaryServers            []string                              `mapstructure:"primary_servers,omitempty"`
	BufferSize                ProtocolsFpolicyExternalEngineBufferSize `mapstructure:"buffer_size,omitempty"`
	SecondaryServers          []string                              `mapstructure:"secondary_servers,omitempty"`
	Port                      int64                                 `mapstructure:"port,omitempty"`
	ServerProgressTimeout     string                                `mapstructure:"server_progress_timeout,omitempty"`
	Format                    string                                `mapstructure:"format,omitempty"`
	Type                      string                                `mapstructure:"type,omitempty"`
	Resiliency                ProtocolsFpolicyExternalEngineResiliency `mapstructure:"resiliency,omitempty"`
	MaxServerRequests         int64                                 `mapstructure:"max_server_requests,omitempty"`
	MaxConnectionRetries      int64                                 `mapstructure:"max_connection_retries,omitempty"`
}

// ProtocolsFpolicyExternalEngineDataSourceFilterModel describes the data source data model for queries.
type ProtocolsFpolicyExternalEngineDataSourceFilterModel struct {
	Name     string `mapstructure:"name"`
	SVMName  string `mapstructure:"svm.name"`
	SVMUUID  string `mapstructure:"svm.uuid"`
}

// GetProtocolsFpolicyExternalEngineByName to get protocols_fpolicy_external_engine info
func GetProtocolsFpolicyExternalEngineByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmUUID string) (*ProtocolsFpolicyExternalEngineGetDataModelONTAP, error) {
	api := fmt.Sprintf("protocols/fpolicy/%s/engines/%s", svmUUID, name)
	query := r.NewQuery()
	query.Fields([]string{"name", "svm", "keep_alive_interval", "request_cancel_timeout", "certificate", "session_timeout", "request_abort_timeout", "status_request_interval", "ssl_option", "primary_servers", "buffer_size", "secondary_servers", "port", "server_progress_timeout", "format", "type", "resiliency", "max_server_requests", "max_connection_retries"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_fpolicy_external_engine info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ProtocolsFpolicyExternalEngineGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_fpolicy_external_engine data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetProtocolsFpolicyExternalEngines to get protocols_fpolicy_external_engine info for all resources matching a filter
func GetProtocolsFpolicyExternalEngines(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *ProtocolsFpolicyExternalEngineDataSourceFilterModel) ([]ProtocolsFpolicyExternalEngineGetDataModelONTAP, error) {
	api := fmt.Sprintf("protocols/fpolicy/%s/engines", filter.SVMUUID)
	query := r.NewQuery()
	query.Fields([]string{"name", "svm", "keep_alive_interval", "request_cancel_timeout", "certificate", "session_timeout", "request_abort_timeout", "status_request_interval", "ssl_option", "primary_servers", "buffer_size", "secondary_servers", "port", "server_progress_timeout", "format", "type", "resiliency", "max_server_requests", "max_connection_retries"})
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding protocols_fpolicy_external_engines filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_fpolicy_external_engines info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []ProtocolsFpolicyExternalEngineGetDataModelONTAP
	for _, info := range response {
		var record ProtocolsFpolicyExternalEngineGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_fpolicy_external_engines data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateProtocolsFpolicyExternalEngine to create protocols_fpolicy_external_engine
func CreateProtocolsFpolicyExternalEngine(errorHandler *utils.ErrorHandler, r restclient.RestClient, body ProtocolsFpolicyExternalEngineResourceBodyDataModelONTAP, svmUUID string) (*ProtocolsFpolicyExternalEngineGetDataModelONTAP, error) {
	api := fmt.Sprintf("protocols/fpolicy/%s/engines", svmUUID)
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding protocols_fpolicy_external_engine body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating protocols_fpolicy_external_engine", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ProtocolsFpolicyExternalEngineGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding protocols_fpolicy_external_engine info", fmt.Sprintf("error on decode storage/protocols_fpolicy_external_engines info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create protocols_fpolicy_external_engine source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// UpdateProtocolsFpolicyExternalEngine to update protocols_fpolicy_external_engine
func UpdateProtocolsFpolicyExternalEngine(errorHandler *utils.ErrorHandler, r restclient.RestClient, body ProtocolsFpolicyExternalEngineResourceBodyDataModelONTAP, svmUUID string, name string) error {
	api := fmt.Sprintf("protocols/fpolicy/%s/engines/%s", svmUUID, name)
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return errorHandler.MakeAndReportError("error encoding protocols_fpolicy_external_engine body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	statusCode, _, err := r.CallUpdateMethod(api, nil, bodyMap)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating protocols_fpolicy_external_engine", fmt.Sprintf("error on PATCH %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}

// DeleteProtocolsFpolicyExternalEngine to delete protocols_fpolicy_external_engine
func DeleteProtocolsFpolicyExternalEngine(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmUUID string, name string) error {
	api := fmt.Sprintf("protocols/fpolicy/%s/engines/%s", svmUUID, name)
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting protocols_fpolicy_external_engine", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
