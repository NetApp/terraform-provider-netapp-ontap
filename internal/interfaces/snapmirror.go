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
