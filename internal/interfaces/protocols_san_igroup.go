package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// TODO:
// copy this file to match you data source (should match internal/interfaces/protocols_san_igroup.go)
// replace ProtocolsSanIgroup with the name of the resource, following go conventions, eg IPInterface
// replace protocols_san_igroup with the name of the resource, for logging purposes, eg ip_interface
// replace api_url with API, eg ip/interfaces
// delete these 5 lines

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

type IgroupsLunMap struct {
	LogicalUnitNumber int                `mapstructure:"logical_unit_number"`
	Lun               IgroupLunForLunMap `mapstructure:"lun"`
}

type IgroupLunForLunMap struct {
	Name string `mapstructure:"name"`
	UUID string `mapstructure:"uuid"`
}

type IgroupLun struct {
	Name    string `mapstructure:"name"`
	UUID    string `mapstructure:"uuid"`
	Comment string `mapstructure:"comment"`
}

type IgroupInitiator struct {
	Name    string `mapstructure:"name"`
	Comment string `mapstructure:"comment"`
}

type Portset struct {
	Name string `mapstructure:"name"`
	UUID string `mapstructure:"uuid"`
}

// ProtocolsSanIgroupResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type ProtocolsSanIgroupResourceBodyDataModelONTAP struct {
	Name string `mapstructure:"name"`
	SVM  svm    `mapstructure:"svm"`
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
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_san_igroup data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

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
