package interfaces

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
)

// ClusterGetDataModelONTAP describes the GET record data model using go types for mapping.
type ClusterGetDataModelONTAP struct {
	// ConfigurableAttribute types.String `json:"configurable_attribute"`
	// ID                    types.String `json:"id"`
	Name    string
	Version versionModelONTAP
}

type versionModelONTAP struct {
	Full string
}

type ipAddress struct {
	Address string
}

type mgmtInterface struct {
	IP ipAddress
}

// ClusterNodeGetDataModelONTAP describes the GET record data model using go types for mapping.
type ClusterNodeGetDataModelONTAP struct {
	// ConfigurableAttribute types.String `json:"configurable_attribute"`
	// ID                    types.String `json:"id"`
	Name                 string
	ManagementInterfaces []mgmtInterface `mapstructure:"management_interfaces"`
	// Version versionModelONTAP
}

// GetCluster to get cluster info
func GetCluster(ctx context.Context, diags diag.Diagnostics, r restclient.RestClient) (*ClusterGetDataModelONTAP, error) {
	statusCode, response, err := r.GetNilOrOneRecord("cluster", nil, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET cluster")
	}
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Read cluster data - error: %s", err))
		// TODO: diags.Error is not reporting anything here.  Works in the caller.
		diags.AddError(err.Error(), fmt.Sprintf("statusCode %d, error %s", statusCode, err))
		return nil, err
	}

	var dataONTAP ClusterGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		tflog.Error(ctx, fmt.Sprintf("Read cluster data - decode error: %s, data: %#v", err, response))
		diags.AddError("failed to unmarshall response from GET cluster - UDATA", fmt.Sprintf("statusCode %d, response %#v", statusCode, response))
		return nil, err
	}
	tflog.Debug(ctx, fmt.Sprintf("Read cluster data source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetClusterNodes to get cluster nodes info
func GetClusterNodes(ctx context.Context, diags diag.Diagnostics, r restclient.RestClient) ([]ClusterNodeGetDataModelONTAP, error) {

	query := r.NewQuery()
	query.Fields([]string{"management_interfaces", "name"})

	statusCode, records, err := r.GetZeroOrMoreRecords("cluster/nodes", query, nil)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Read cluster nodes data - error: %s", err))
		diags.AddError(err.Error(), fmt.Sprintf("statusCode %d, result %#v", statusCode, records))
		return nil, err
	}
	tflog.Debug(ctx, fmt.Sprintf("Read cluster data source NODES - records: %#v", records))

	var dataONTAP ClusterNodeGetDataModelONTAP
	nodes := []ClusterNodeGetDataModelONTAP{}
	for _, record := range records {
		if err := mapstructure.Decode(record, &dataONTAP); err != nil {
			tflog.Error(ctx, fmt.Sprintf("Read cluster node data - decode error: %s", err))
			// TODO: diags.Error is not reporting anything here.  Works in the caller.
			diags.AddError("failed to unmarshall response from GET cluster - UDATA", fmt.Sprintf("statusCode %d, result %#v", statusCode, records))
			return nil, err
		}
		nodes = append(nodes, dataONTAP)
	}
	tflog.Debug(ctx, fmt.Sprintf("Read cluster data source NODES - udata: %#v", nodes))
	return nodes, nil
}
