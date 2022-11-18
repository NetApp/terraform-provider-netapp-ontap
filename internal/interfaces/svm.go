package interfaces

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
)

// SvmGetDataModelONTAP describes the GET record data model using go types for mapping.
type SvmGetDataModelONTAP struct {
	Name string
	UUID string
}

// SvmResourceModel describes the resource data model.
type SvmResourceModel struct {
	Name string `mapstructure:"name"`
}

// GetSvm to get svm info by uuid
func GetSvm(ctx context.Context, diags diag.Diagnostics, r restclient.RestClient, uuid string) (*SvmGetDataModelONTAP, error) {
	statusCode, response, err := r.GetNilOrOneRecord("svm/svms/"+uuid, nil, nil)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Read vserver data - error: %s", err))
		// TODO: diags.Error is not reporting anything here.  Works in the caller.
		diags.AddError(err.Error(), fmt.Sprintf("statusCode %d, error %s", statusCode, err))
		return nil, err
	}

	var dataONTAP *SvmGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		tflog.Error(ctx, fmt.Sprintf("Read vserver data - decode error: %s, data: %#v", err, response))
		diags.AddError("failed to get vserver response from GET svm/svms/ - UDATA", fmt.Sprintf("statusCode %d, response %#v", statusCode, response))
		return nil, err
	}
	tflog.Debug(ctx, fmt.Sprintf("Read vserver source - udata: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateSvm to create vserver
func CreateSvm(ctx context.Context, diags diag.Diagnostics, r restclient.RestClient, data SvmResourceModel) (*SvmGetDataModelONTAP, error) {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		tflog.Error(ctx, fmt.Sprintf("Create vserver - encode error: %s, data: %#v", err, data))
		return nil, err
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod("svm/svms", query, body)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Create vserver - error: %s", err))
		// TODO: diags.Error is not reporting anything here.  Works in the caller.
		diags.AddError(err.Error(), fmt.Sprintf("statusCode %d, error %s", statusCode, err))
		return nil, err
	}

	var dataONTAP SvmGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		tflog.Error(ctx, fmt.Sprintf("Create vserver - decode error: %s, data: %#v", err, response))
		diags.AddError("failed to get response from POST svm/svms/ - UDATA", fmt.Sprintf("statusCode %d, response %#v", statusCode, response))
		return nil, err
	}
	tflog.Debug(ctx, fmt.Sprintf("Create vserver source - udata: %#v", dataONTAP))
	return &dataONTAP, nil

}

// DeleteSvm to delete vserver
func DeleteSvm(ctx context.Context, diags diag.Diagnostics, r restclient.RestClient, uuid string) error {
	statusCode, _, err := r.CallDeleteMethod("svm/svms/"+uuid, nil, nil)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Delete vserver - error: %s", err))
		// TODO: diags.Error is not reporting anything here.  Works in the caller.
		diags.AddError(err.Error(), fmt.Sprintf("statusCode %d, error %s", statusCode, err))
		return err
	}

	return nil

}
