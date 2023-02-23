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
	Cron     CronSchedule `mapstructure:"cron",omitempty`
	Interval string       `mapstructure:"interval",omitempty`
}

// CronSchedule is the body data model for cron schedule fields
type CronSchedule struct {
	Hours    []int64 `mapstructure:"hours",omitempty`
	Days     []int64 `mapstructure:"days",omitempty`
	Minutes  []int64 `mapstructure:"minutes",omitempty`
	Weekdays []int64 `mapstructure:"weekdays",omitempty`
	Months   []int64 `mapstructure:"momths",omitempty`
}

// GetClusterSchedule to get a single schedule info
func GetClusterSchedule(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string) (*ClusterScheduleGetDataModelONTAP, error) {
	query := r.NewQuery()
	query.Set("name", name)
	api := "cluster/schedules"
	query.Fields([]string{"name", "type", "uuid", "cron", "interval", "scope"})

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading schedlue info",
			fmt.Sprintf("error on GET %s: %s, statuscode: %d", api, err, statusCode))
	}

	if response == nil {
		tflog.Debug(errorHandler.Ctx, fmt.Sprintf("schedule %s not found", name))
		return nil, nil
	}

	var dataONTAP ClusterScheduleGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding schedule info",
			fmt.Sprintf("statusCode %d, response %#v", statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read cluster/schedules data source: %#v", dataONTAP))
	return &dataONTAP, nil
}
