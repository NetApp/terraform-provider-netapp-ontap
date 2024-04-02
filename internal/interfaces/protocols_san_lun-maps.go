package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// ProtocolsSanLunMapsGetDataModelONTAP describes the GET record data model using go types for mapping.
type ProtocolsSanLunMapsGetDataModelONTAP struct {
	SVM               svm    `mapstructure:"svm"`
	LogicalUnitNumber int    `mapstructure:"logical_unit_number"`
	IGroup            IGroup `mapstructure:"igroup"`
	Lun               Lun    `mapstructure:"lun"`
}

// igroup describes the resource data model.
type IGroup struct {
	Name string `mapstructure:"name,omitempty"`
	UUID string `mapstructure:"uuid,omitempty"`
}

// lun describes the resource data model.
type Lun struct {
	Name string `mapstructure:"name,omitempty"`
	UUID string `mapstructure:"uuid,omitempty"`
}

// ProtocolsSanLunMapsResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type ProtocolsSanLunMapsResourceBodyDataModelONTAP struct {
	SVM               svm    `mapstructure:"svm"`
	IGroup            IGroup `mapstructure:"igroup"`
	Lun               Lun    `mapstructure:"lun"`
	LogicalUnitNumber int    `mapstructure:"logical_unit_number,omitempty"`
}

// ProtocolsSanLunMapsDataSourceFilterModel describes the data source data model for queries.
type ProtocolsSanLunMapsDataSourceFilterModel struct {
	Lun    Lun    `mapstructure:"lun"`
	SVM    SVM    `mapstructure:"svm"`
	IGroup IGroup `mapstructure:"igroup"`
}

// GetProtocolsSanLunMaps to get protocols_san_lun-maps info
func GetProtocolsSanLunMaps(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *ProtocolsSanLunMapsDataSourceFilterModel) ([]ProtocolsSanLunMapsGetDataModelONTAP, error) {
	api := "/protocols/san/lun-maps"
	query := r.NewQuery()
	query.Fields([]string{"svm.name", "igroup.name", "igroup.uuid", "lun.name", "lun.uuid", "logical_unit_number"})
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding /protocols/san/lun-maps filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}

	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_san_lun-maps info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []ProtocolsSanLunMapsGetDataModelONTAP
	for _, info := range response {
		var record ProtocolsSanLunMapsGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_san_lun-maps data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// GetProtocolsSanLunMapsByName to get protocols_san_lun-maps info
func GetProtocolsSanLunMapsByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, igroupName string, lunName string, svmName string) (*ProtocolsSanLunMapsGetDataModelONTAP, error) {
	api := "/protocols/san/lun-maps"
	query := r.NewQuery()
	query.Set("igroup.name", igroupName)
	query.Set("lun.name", lunName)
	query.Set("svm.name", svmName)
	query.Fields([]string{"svm.name", "igroup.name", "igroup.uuid", "lun.name", "lun.uuid", "logical_unit_number"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_san_lun-maps info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ProtocolsSanLunMapsGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_san_lun-maps data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// CreateProtocolsSanLunMaps to create protocols_san_lun-maps
func CreateProtocolsSanLunMaps(errorHandler *utils.ErrorHandler, r restclient.RestClient, body ProtocolsSanLunMapsResourceBodyDataModelONTAP) (*ProtocolsSanLunMapsGetDataModelONTAP, error) {
	api := "/protocols/san/lun-maps"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding protocols_san_lun-maps body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating protocols_san_lun-maps", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ProtocolsSanLunMapsGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding protocols_san_lun-maps info", fmt.Sprintf("error on decode storage/protocols_san_lun-mapss info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create protocols_san_lun-maps source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteProtocolsSanLunMaps to delete protocols_san_lun-maps
func DeleteProtocolsSanLunMaps(errorHandler *utils.ErrorHandler, r restclient.RestClient, igroupUUID string, lunUUID string) error {
	api := fmt.Sprintf("/protocols/san/lun-maps/%s/%s", lunUUID, igroupUUID)
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting protocols_san_lun-maps", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
