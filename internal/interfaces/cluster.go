package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// ClusterGetDataModelONTAP describes the GET record data model using go types for mapping.
type ClusterGetDataModelONTAP struct {
	// ConfigurableAttribute types.String `json:"configurable_attribute"`
	// ID                    types.String `json:"id"`
	Name                 string
	Version              versionModelONTAP
	Contact              string
	Location             string
	DnsDomains           []string           `mapstructure:"dns_domains"`
	NameServers          []string           `mapstructure:"name_servers"`
	NtpServers           []string           `mapstructure:"ntp_servers"`
	TimeZone             timeZone           `mapstructure:"timezone"`
	ClusterCertificate   ClusterCertificate `mapstructure:"certificate"`
	ManagementInterfaces []mgmtInterface    `mapstructure:"management_interfaces"`
}

type timeZone struct {
	Name string
}

type mgmtInterface struct {
	IP   ipAddress `mapstructure:"ip"`
	Name string    `mapstructure:"name"`
	ID   string    `mapstructure:"uuid"`
}

type ClusterCertificate struct {
	ID string `mapstructure:"uuid"`
}

type versionModelONTAP struct {
	Full       string
	Generation int
	Major      int
	Minor      int
}

type ipAddress struct {
	Address string
}

type noddMgmtInterface struct {
	IP ipAddress
}

// ClusterNodeGetDataModelONTAP describes the GET record data model using go types for mapping.
type ClusterNodeGetDataModelONTAP struct {
	Name                 string
	ManagementInterfaces []noddMgmtInterface `mapstructure:"management_interfaces"`
	// Version versionModelONTAP
}

// GetCluster to get cluster info
func GetCluster(errorHandler *utils.ErrorHandler, r restclient.RestClient) (*ClusterGetDataModelONTAP, error) {
	statusCode, response, err := r.GetNilOrOneRecord("cluster", nil, nil)
	query := r.NewQuery()
	query.Fields([]string{"name", "location", "contact", "dns_domains", "name_servers", "ntp_servers", "management_interfaces", "timezone", "certificate"})
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET cluster")
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading cluster info", fmt.Sprintf("error on GET cluster: %s, statusCode %d", err, statusCode))
	}

	var dataONTAP ClusterGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("failed to decode response from GET cluster", fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read cluster data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetClusterNodes to get cluster nodes info
func GetClusterNodes(errorHandler *utils.ErrorHandler, r restclient.RestClient) ([]ClusterNodeGetDataModelONTAP, error) {

	query := r.NewQuery()
	query.Fields([]string{"management_interfaces", "name"})

	statusCode, records, err := r.GetZeroOrMoreRecords("cluster/nodes", query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading cluster nodes info", fmt.Sprintf("error on GET cluster/nodes: %s", err))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read cluster data source NODES - records: %#v", records))

	var dataONTAP ClusterNodeGetDataModelONTAP
	nodes := []ClusterNodeGetDataModelONTAP{}
	for _, record := range records {
		if err := mapstructure.Decode(record, &dataONTAP); err != nil {
			return nil, errorHandler.MakeAndReportError("error decoding cluster nodes info", fmt.Sprintf("error: %s, statusCode %d, record %#v", err, statusCode, record))
		}
		nodes = append(nodes, dataONTAP)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read cluster data source NODES: %#v", nodes))
	return nodes, nil
}
