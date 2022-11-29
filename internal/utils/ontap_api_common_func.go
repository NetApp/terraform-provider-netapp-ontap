package utils

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
)

// GetJobByID returns the job state given the job uuid.
func GetJobByID(errorHandler *ErrorHandler, r restclient.RestClient, uuid string) (interface{}, error) {
	api := "cluster/jobs/" + uuid
	statusCode, record, err := r.GetNilOrOneRecord(api, nil, nil)
	if err == nil && record == nil {
		err = fmt.Errorf("no response for GET job")
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading job info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))

	}
	var job jobModel
	if err := mapstructure.Decode(record, &job); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding job info", fmt.Sprintf("error: %s, statusCode %d, record %#v", err, statusCode, record))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read job: %#v", job))
	return &job, nil
}

type jobModel struct {
	State string `tfsdk:"state"`
}
