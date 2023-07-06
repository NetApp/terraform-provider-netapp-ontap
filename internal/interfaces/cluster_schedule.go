package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// ClusterScheduleGetDataModelONTAP describes the GET record data model using go types for mapping.
type ClusterScheduleGetDataModelONTAP struct {
	Name     string       `mapstructure:"name"`
	UUID     string       `mapstructure:"uuid"`
	Type     string       `mapstructure:"type"`
	Scope    string       `mapstructure:"scope"`
	Cron     CronSchedule `mapstructure:"cron,omitempty"`
	Interval string       `mapstructure:"interval,omitempty"`
}

// CronSchedule is the body data model for cron schedule fields
type CronSchedule struct {
	Hours    []int64 `mapstructure:"hours,omitempty"`
	Days     []int64 `mapstructure:"days,omitempty"`
	Minutes  []int64 `mapstructure:"minutes,omitempty"`
	Weekdays []int64 `mapstructure:"weekdays,omitempty"`
	Months   []int64 `mapstructure:"months,omitempty"`
}

// ClusterScheduleResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type ClusterScheduleResourceBodyDataModelONTAP struct {
	// 'name' is not allowed in the API body for the update. Set omitempty to keep the flexibility.
	Name     string       `mapstructure:"name,omitempty"`
	Cron     CronSchedule `mapstructure:"cron,omitempty"`
	Interval string       `mapstructure:"interval,omitempty"`
}

// GetClusterSchedule to get a single schedule info
func GetClusterSchedule(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string) (*ClusterScheduleGetDataModelONTAP, error) {
	query := r.NewQuery()
	query.Set("name", name)
	api := "cluster/schedules"
	query.Fields([]string{"name", "uuid", "cron", "interval", "type", "scope"})

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading schedule info",
			fmt.Sprintf("error on GET %s: %s, statuscode: %d", api, err, statusCode))
	}

	if response == nil {
		tflog.Debug(errorHandler.Ctx, fmt.Sprintf("schedule %s not found", name))
		return nil, errorHandler.MakeAndReportError("error reading schedule info", fmt.Sprintf("schedule %s not found", name))
	}

	var dataONTAP ClusterScheduleGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding schedule info",
			fmt.Sprintf("statusCode %d, response %#v", statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read cluster/schedules data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// CreateClusterSchedule to create job schedule
func CreateClusterSchedule(errorHandler *utils.ErrorHandler, r restclient.RestClient, body ClusterScheduleResourceBodyDataModelONTAP) (*ClusterScheduleGetDataModelONTAP, error) {
	api := "cluster/schedules"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding cluster_schedules body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating cluster_schedule", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ClusterScheduleGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding cluster_schedule info", fmt.Sprintf("error on decode cluster/schedule info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create cluster_schedule source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// UpdateClusterSchedule to update a job schedule
func UpdateClusterSchedule(errorHandler *utils.ErrorHandler, r restclient.RestClient, data ClusterScheduleResourceBodyDataModelONTAP, id string) error {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return errorHandler.MakeAndReportError("error encoding clustser schedule body", fmt.Sprintf("error on encoding cluster schedule body: %s, body: %#v", err, data))
	}

	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, _, err := r.CallUpdateMethod("cluster/schedules/"+id, query, body)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating cluster schedule", fmt.Sprintf("error on POST cluster/schedules: %s, statusCode %d", err, statusCode))
	}
	return nil

}

// DeleteClusterSchedule to delete job schedule
func DeleteClusterSchedule(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := "cluster/schedules"
	statusCode, _, err := r.CallDeleteMethod(api+"/"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting cluster_schedule", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
