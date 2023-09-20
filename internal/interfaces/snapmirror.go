package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// SnapmirrorGetDataModelONTAP defines the resource get data model
type SnapmirrorGetDataModelONTAP struct {
	Healthy bool   `mapstructure:"healthy"`
	State   string `mapstructure:"state"`
	UUID    string `mapstructure:"uuid"`
}

// SnapmirrorGetRawDataModelONTAP defines the resource get data model
type SnapmirrorGetRawDataModelONTAP struct {
	UUID string `mapstructure:"uuid"`
}

// SnapmirrorResourceBodyDataModelONTAP defines the resource data model
type SnapmirrorResourceBodyDataModelONTAP struct {
	SourceEndPoint      EndPoint          `mapstructure:"source"`
	DestinationEndPoint EndPoint          `mapstructure:"destination"`
	CreateDestination   CreateDestination `mapstructure:"create_destination,omitempty"`
}

// EndPoint defines source/destination endpoint data model.
type EndPoint struct {
	Cluster Cluster `mapstructure:"cluster,omitempty"`
	Path    string  `mapstructure:"path"`
}

// CreateDestination defines CreateDestination data model.
type CreateDestination struct {
	Enabled bool `mapstructure:"enabled"`
}

// Cluster defines Cluster data model.
type Cluster struct {
	Name string `mapstructure:"name,omitempty"`
}

// SnapmirrorFilterModel Snapmirror filter model
type SnapmirrorFilterModel struct {
	DestinationPath string `mapstructure:"destination.path"`
}

// SnapmirrorDataSourceModel data model
type SnapmirrorDataSourceModel struct {
	Source      Source           `mapstructure:"source"`
	Destination Destination      `mapstructure:"destination"`
	Healthy     bool             `mapstructure:"healthy"`
	Restore     bool             `mapstructure:"restore"`
	UUID        string           `mapstructure:"uuid"`
	State       string           `mapstructure:"state"`
	Policy      SnapmirrorPolicy `mapstructure:"policy"`
	GroupType   string           `mapstructure:"group_type"`
	Throttle    int              `mapstructure:"throttle"`
}

// Source data model
type Source struct {
	Cluster SnapmirrorCluster `mapstructure:"cluster"`
	Path    string            `mapstructure:"path"`
	Svm     SvmDataModelONTAP `mapstructure:"svm"`
}

// Destination data model
type Destination struct {
	Path string            `mapstructure:"path"`
	Svm  SvmDataModelONTAP `mapstructure:"svm"`
}

// SnapmirrorCluster data model
type SnapmirrorCluster struct {
	Name string `mapstructure:"name"`
	UUID string `mapstructure:"uuid"`
}

// SnapmirrorPolicy data model
type SnapmirrorPolicy struct {
	UUID string `mapstructure:"uuid"`
}

// GetSnapmirrorByID ...
func GetSnapmirrorByID(errorHandler *utils.ErrorHandler, r restclient.RestClient, id string) (*SnapmirrorGetDataModelONTAP, error) {
	api := "snapmirror/relationships/" + id
	statusCode, response, err := r.GetNilOrOneRecord(api, nil, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading snapmirror info", fmt.Sprintf("error on GET %s: %s", api, err))
	}
	var rawDataONTAP SnapmirrorGetDataModelONTAP
	if err := mapstructure.Decode(response, &rawDataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding snapmirror info", fmt.Sprintf("error on decode %s: %s, statusCode %d, response %#v", api, err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read snapmirror source - udata: %#v", rawDataONTAP))
	return &rawDataONTAP, nil
}

// GetSnapmirrorByDestinationPath to get snapmirror data source info by Destination Path
func GetSnapmirrorByDestinationPath(errorHandler *utils.ErrorHandler, r restclient.RestClient, destinationPath string, version versionModelONTAP) (*SnapmirrorDataSourceModel, error) {
	api := "snapmirror/relationships"
	query := r.NewQuery()
	query.Add("destination.path", destinationPath)
	fields := []string{"destination", "healthy", "source", "restore", "policy", "state"}
	if version.Generation == 9 && version.Major > 10 {
		fields = append(fields, "throttle", "group_type")
	}
	query.Fields(fields)

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading snapmirror/relationships info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP SnapmirrorDataSourceModel

	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read snapmirror/relationships data source: %#v", dataONTAP))

	return &dataONTAP, nil
}

// GetSnapmirrors to get list of policies
func GetSnapmirrors(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *SnapmirrorFilterModel, version versionModelONTAP) ([]SnapmirrorDataSourceModel, error) {
	api := "snapmirror/relationships"
	query := r.NewQuery()

	fields := []string{"unhealthy_reason", "destination", "healthy", "source", "restore", "policy", "transfer", "state", "exported_snapshot", "lag_time"}
	if version.Generation == 9 && version.Major > 7 {
		fields = append(fields, "consistency_group_failover")
	}
	if version.Generation == 9 && version.Major > 10 {
		fields = append(fields, "identity_preservation", "throttle", "transfer_schedule", "group_type", "last_transfer_type")
	}
	if version.Generation == 9 && version.Major > 12 {
		fields = append(fields, "total_transfer_duration", "last_transfer_network_compression_ratio", "total_transfer_bytes", "svmdr_volumes")
	}
	query.Fields(fields)
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding snapmirror/relationships filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading snapmirror/relationships info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []SnapmirrorDataSourceModel
	for _, info := range response {
		var record SnapmirrorDataSourceModel
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read snapmirror/relationships data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateSnapmirror to create snapmirror
func CreateSnapmirror(errorHandler *utils.ErrorHandler, r restclient.RestClient, body SnapmirrorResourceBodyDataModelONTAP) (*SnapmirrorGetRawDataModelONTAP, error) {
	api := "snapmirror/relationships"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding snapmirror/relationships body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	// tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read vserver info: %#v", bodyMap))
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating snapmirror", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var rawDataONTAP SnapmirrorGetRawDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &rawDataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding snapmirror info", fmt.Sprintf("error on decode snapmirror info: %s, statusCode %d, response %#v", err, statusCode, response))
	}

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create snapmirror source - udata: %#v", rawDataONTAP))
	return &rawDataONTAP, nil
}

// InitializeSnapmirror ...
func InitializeSnapmirror(errorHandler *utils.ErrorHandler, r restclient.RestClient, id string, state string) error {
	api := "snapmirror/relationships/" + id
	body := map[string]interface{}{"state": state}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallUpdateMethod(api, query, body)
	if err != nil {
		return errorHandler.MakeAndReportError("error initializing snapmirror", fmt.Sprintf("error on PATCH %s: %s, statusCode %d, response %#v", api, err, statusCode, response))
	}

	return nil
}

// DeleteSnapmirror to delete ip_interface
func DeleteSnapmirror(errorHandler *utils.ErrorHandler, r restclient.RestClient, id string) error {
	api := "snapmirror/relationships/" + id
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting snapmirror/relationships", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
