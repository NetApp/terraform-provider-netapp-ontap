package utils

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
)

// GetJobByID returns the job state given the job uuid.
func GetJobByID(ctx context.Context, diags diag.Diagnostics, r restclient.RestClient, uuid string) (interface{}, error) {
	statusCode, response, err := r.GetNilOrOneRecord("cluster/jobs/"+uuid, nil, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET job")
	}
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Read job data - error: %s", err))
		// TODO: diags.Error is not reporting anything here.  Works in the caller.
		diags.AddError(err.Error(), fmt.Sprintf("statusCode %d, error %s", statusCode, err))
		return nil, err
	}
	var job jobModel
	if err := mapstructure.Decode(response, &job); err != nil {
		tflog.Error(ctx, fmt.Sprintf("Read job data - decode error: %s, data: %#v", err, response))
		diags.AddError("failed to unmarshall response from GET job", fmt.Sprintf("statusCode %d, response %#v", statusCode, response))
		return nil, err
	}
	tflog.Debug(ctx, fmt.Sprintf("Read job: %#v", job))
	return &job, nil
}

type jobModel struct {
	State string `tfsdk:"state"`
}
