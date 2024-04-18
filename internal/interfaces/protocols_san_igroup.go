package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// ProtocolsSanIgroupGetDataModelONTAP describes the GET record data model using go types for mapping.
type ProtocolsSanIgroupGetDataModelONTAP struct {
	Name       string            `mapstructure:"name"`
	UUID       string            `mapstructure:"uuid"`
	SVM        SvmDataModelONTAP `mapstructure:"svm"`
	LunMaps    []IgroupsLunMap   `mapstructure:"lun_maps"`
	OsType     string            `mapstructure:"os_type"`
	Protocol   string            `mapstructure:"protocol"`
	Comment    string            `mapstructure:"comment"`
	Igroups    []IgroupLun       `mapstructure:"igroups"`
	Initiators []IgroupInitiator `mapstructure:"initiators"`
	Portset    Portset           `mapstructure:"portset"`
}

// IgroupsLunMap describes the data model for lun_maps.
type IgroupsLunMap struct {
	LogicalUnitNumber int                `mapstructure:"logical_unit_number"`
	Lun               IgroupLunForLunMap `mapstructure:"lun"`
}

// IgroupLunForLunMap describes the data model for lun.
type IgroupLunForLunMap struct {
	Name string `mapstructure:"name"`
	UUID string `mapstructure:"uuid"`
}

// IgroupLun describes the data model for igroup.
type IgroupLun struct {
	Name    string `mapstructure:"name"`
	UUID    string `mapstructure:"uuid,omitempty"`
	Comment string `mapstructure:"comment,omitempty"`
}

// IgroupInitiator describes the data model for initiator.
type IgroupInitiator struct {
	Name    string `mapstructure:"name"`
	Comment string `mapstructure:"comment"`
}

// Portset describes the data model for portset.
type Portset struct {
	Name string `mapstructure:"name,omitempty"`
	UUID string `mapstructure:"uuid,omitempty"`
}

// ProtocolsSanIgroupResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type ProtocolsSanIgroupResourceBodyDataModelONTAP struct {
	Name       string            `mapstructure:"name"`
	SVM        SvmDataModelONTAP `mapstructure:"svm"`
	OsType     string            `mapstructure:"os_type"`
	Protocol   string            `mapstructure:"protocol"`
	Comment    string            `mapstructure:"comment,omitempty"`
	Igroups    []IgroupLun       `mapstructure:"igroups,omitempty"`
	Initiators []IgroupInitiator `mapstructure:"initiators,omitempty"`
	Portset    Portset           `mapstructure:"portset,omitempty"`
}

// UpdateProtocolsSanIgroupResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type UpdateProtocolsSanIgroupResourceBodyDataModelONTAP struct {
	Name    string `mapstructure:"name,omitempty"`
	OsType  string `mapstructure:"os_type,omitempty"`
	Comment string `mapstructure:"comment,omitempty"`
}

// ProtocolsSanIgroupDataSourceFilterModel describes the data source data model for queries.
type ProtocolsSanIgroupDataSourceFilterModel struct {
	Name    string `mapstructure:"name"`
	SVMName string `mapstructure:"svm.name"`
}

// GetProtocolsSanIgroupByName to get protocols_san_igroup info
func GetProtocolsSanIgroupByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmName string, version versionModelONTAP) (*ProtocolsSanIgroupGetDataModelONTAP, error) {
	api := "protocols/san/igroups"
	query := r.NewQuery()
	query.Set("name", name)
	query.Set("svm.name", svmName)
	var fields = []string{"name", "svm.name", "lun_maps", "os_type", "protocol", "uuid"}
	if version.Generation == 9 && version.Major >= 9 {
		fields = append(fields, "comment", "igroups", "initiators", "portset")
	}
	query.Fields(fields)
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_san_igroup info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ProtocolsSanIgroupGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_san_igroup: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetProtocolsSanIgroups to get protocols_san_igroup info for all resources matching a filter
func GetProtocolsSanIgroups(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *ProtocolsSanIgroupDataSourceFilterModel, version versionModelONTAP) ([]ProtocolsSanIgroupGetDataModelONTAP, error) {
	api := "protocols/san/igroups"
	query := r.NewQuery()
	var fields = []string{"name", "svm.name", "lun_maps", "os_type", "protocol", "uuid"}
	if version.Generation == 9 && version.Major >= 9 {
		fields = append(fields, "comment", "igroups", "initiators", "portset")
	}
	if filter != nil {
		if filter.Name != "" {
			query.Add("name", filter.Name)
		}
		if filter.SVMName != "" {
			query.Add("svm.name", filter.SVMName)
		}
	}
	query.Fields(fields)
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_san_igroup info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []ProtocolsSanIgroupGetDataModelONTAP
	for _, info := range response {
		var record ProtocolsSanIgroupGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)

	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_san_igroup data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateProtocolsSanIgroup to create protocols_san_igroup
func CreateProtocolsSanIgroup(errorHandler *utils.ErrorHandler, r restclient.RestClient, body ProtocolsSanIgroupResourceBodyDataModelONTAP) (*ProtocolsSanIgroupGetDataModelONTAP, error) {
	api := "protocols/san/igroups"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding protocols_san_igroups body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating protocols_san_igroups", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ProtocolsSanIgroupGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding protocols_san_igroups info", fmt.Sprintf("error on decode storage/protocols_san_igroups info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create protocols_san_igroups source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// UpdateProtocolsSanIgroup to update a protocols_san_igroup
func UpdateProtocolsSanIgroup(errorHandler *utils.ErrorHandler, r restclient.RestClient, data UpdateProtocolsSanIgroupResourceBodyDataModelONTAP, uuid string) error {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return errorHandler.MakeAndReportError("error encoding protocols_san_igroup body", fmt.Sprintf("error on encoding protocols/san/igroups body: %s, body: %#v", err, data))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Update protocols_san_igroup info: %#v", data))
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, _, err := r.CallUpdateMethod("protocols/san/igroups/"+uuid, query, body)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating protocols_san_igroup", fmt.Sprintf("error on PATCH protocols/san/igroups: %s, statusCode %d", err, statusCode))
	}
	return nil
}

// DeleteProtocolsSanIgroup to delete protocols_san_igroup
func DeleteProtocolsSanIgroup(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := fmt.Sprintf("protocols/san/igroups/%s", uuid)
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting protocols_san_igroups", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
