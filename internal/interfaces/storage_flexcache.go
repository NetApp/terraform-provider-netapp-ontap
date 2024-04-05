package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// StorageFlexcacheGetDataModelONTAP describes the GET record data model using go types for mapping.
type StorageFlexcacheGetDataModelONTAP struct {
	Name                     string
	SVM                      svm
	Aggregates               []StorageFlexcacheAggregate `mapstructure:"aggregates"`
	Origins                  []StorageFlexcacheOrigin    `mapstructure:"origins"`
	JunctionPath             string                      `mapstructure:"junction_path,omitempty"`
	Size                     int                         `mapstructure:"size,omitempty"`
	Path                     string                      `mapstructure:"path,omitempty"`
	Guarantee                StorageFlexcacheGuarantee   `mapstructure:"guarantee,omitempty"`
	DrCache                  bool                        `mapstructure:"dr_cache,omitempty"`
	GlobalFileLockingEnabled bool                        `mapstructure:"global_file_locking_enabled,omitempty"`
	UseTieredAggregate       bool                        `mapstructure:"use_tiered_aggregate,omitempty"`
	ConstituentsPerAggregate int                         `mapstructure:"constituents_per_aggregate,omitempty"`
	UUID                     string
}

// StorageFlexcacheResourceModel describes the resource data model.
type StorageFlexcacheResourceModel struct {
	Name                     string                    `mapstructure:"name,omitempty"`
	SVM                      svm                       `mapstructure:"svm,omitempty"`
	Origins                  []map[string]interface{}  `mapstructure:"origins,omitempty"`
	JunctionPath             string                    `mapstructure:"junction_path,omitempty"`
	Size                     int                       `mapstructure:"size,omitempty"`
	Path                     string                    `mapstructure:"path,omitempty"`
	Guarantee                StorageFlexcacheGuarantee `mapstructure:"guarantee,omitempty"`
	DrCache                  bool                      `mapstructure:"dr_cache"`
	GlobalFileLockingEnabled bool                      `mapstructure:"global_file_locking_enabled"`
	UseTieredAggregate       bool                      `mapstructure:"use_tiered_aggregate"`
	ConstituentsPerAggregate int                       `mapstructure:"constituents_per_aggregate,omitempty"`
	Aggregates               []map[string]interface{}  `mapstructure:"aggregates,omitempty"`
}

// StorageFlexcacheGuarantee describes the guarantee data model of Guarantee within StorageFlexcacheResourceModel.
type StorageFlexcacheGuarantee struct {
	Type string `mapstructure:"type,omitempty"`
}

// StorageFlexcacheOrigin describes the origin data model of Origin within StorageFlexcacheResourceModel.
type StorageFlexcacheOrigin struct {
	Volume StorageFlexcacheVolume `mapstructure:"volume"`
	SVM    StorageFlexcacheSVM    `mapstructure:"svm"`
}

// StorageFlexcacheVolume describes the volume data model of Volume within StorageFlexcacheOrigin.
type StorageFlexcacheVolume struct {
	Name string `mapstructure:"name,omitempty"`
	ID   string `mapstructure:"uuid,omitempty"`
}

// StorageFlexcacheSVM describes the svm data model of SVM within StorageFlexcacheOrigin.
type StorageFlexcacheSVM struct {
	Name string `mapstructure:"name,omitempty"`
	ID   string `mapstructure:"uuid,omitempty"`
}

// StorageFlexcacheAggregate describes the aggregate data model of Aggregate within StorageFlexcacheResourceModel.
type StorageFlexcacheAggregate struct {
	Name string `mapstructure:"name,omitempty"`
	ID   string `mapstructure:"uuid,omitempty"`
}

// StorageFlexcacheDataSourceFilterModel describes the data source data model for queries.
type StorageFlexcacheDataSourceFilterModel struct {
	Name    string `mapstructure:"name"`
	SVMName string `mapstructure:"svm.name"`
}

// GetStorageFlexcacheByName to get flexcache info by name.
func GetStorageFlexcacheByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmName string) (*StorageFlexcacheGetDataModelONTAP, error) {
	query := r.NewQuery()
	query.Add("name", name)
	query.Add("svm.name", svmName)
	query.Fields([]string{"size", "path", "origins", "guarantee.type", "constituents_per_aggregate", "dr_cache", "global_file_locking_enabled", "use_tiered_aggregate", "aggregates"})
	statusCode, response, err := r.GetNilOrOneRecord("storage/flexcache/flexcaches", query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading flexcache info", fmt.Sprintf("error on GET storage/flexcache/flexcaches: %s", err))
	}
	var dataONTAP *StorageFlexcacheGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding flexcache info", fmt.Sprintf("error on decode storage/flexcache/flexcaches: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read flexcache source - udata: %#v", dataONTAP))
	return dataONTAP, nil
}

// GetStorageFlexcaches to get flexcaches info by filter
func GetStorageFlexcaches(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *StorageFlexcacheDataSourceFilterModel) ([]StorageFlexcacheGetDataModelONTAP, error) {
	api := "storage/flexcache/flexcaches"
	query := r.NewQuery()
	query.Fields([]string{"size", "path", "origins", "guarantee.type", "constituents_per_aggregate", "dr_cache", "global_file_locking_enabled", "use_tiered_aggregate", "aggregates"})
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding storage flexcache filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}

	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading storage flexcache info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []StorageFlexcacheGetDataModelONTAP
	for _, info := range response {
		var record StorageFlexcacheGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read storage flexcache data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateStorageFlexcache creates flexcache.
// POST API returns result, but does not include the attributes that are not set. Make a spearate GET call to get all attributes.
func CreateStorageFlexcache(errorHandler *utils.ErrorHandler, r restclient.RestClient, data StorageFlexcacheResourceModel) error {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return errorHandler.MakeAndReportError("error encoding flexcache body", fmt.Sprintf("error on encoding storage/flexcache/flexcaches body: %s, body: %#v", err, data))
	}
	//The use-tiered-aggregate option is only supported when auto provisioning the FlexCache volume
	if _, ok := body["aggregates"]; ok {
		delete(body, "use_tiered_aggregate")
	}
	query := r.NewQuery()
	query.Add("return_records", "false")
	statusCode, _, err := r.CallCreateMethod("storage/flexcache/flexcaches", query, body)
	if err != nil {
		return errorHandler.MakeAndReportError("error creating flexcache", fmt.Sprintf("error on POST storage/flexcache/flexcaches: %s, statusCode %d", err, statusCode))
	}

	return nil

}

// DeleteStorageFlexcache to delete flexcache by id.
func DeleteStorageFlexcache(errorHandler *utils.ErrorHandler, r restclient.RestClient, id string) error {
	statusCode, _, err := r.CallDeleteMethod("storage/flexcache/flexcaches/"+id, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting flexcache", fmt.Sprintf("error on DELETE storage/flexcache/flexcaches: %s, statusCode %d", err, statusCode))
	}
	return nil
}
