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
	DNSDomains           []string           `mapstructure:"dns_domains"`
	NameServers          []string           `mapstructure:"name_servers"`
	NtpServers           []string           `mapstructure:"ntp_servers"`
	TimeZone             timeZone           `mapstructure:"timezone"`
	ClusterCertificate   ClusterCertificate `mapstructure:"certificate"`
	ManagementInterfaces []mgmtInterface    `mapstructure:"management_interfaces"`
	ID                   string             `mapstructure:"uuid"`
}

// ClusterResourceBodyDataModelONTAP describes the POST/PATCH record data model using go types for mapping.
type ClusterResourceBodyDataModelONTAP struct {
	Name                string               `mapstructure:"name,omitempty"`
	License             ClusterLicense       `mapstructure:"license,omitempty"`
	Contact             string               `mapstructure:"contact,omitempty"`
	Location            string               `mapstructure:"location,omitempty"`
	DNSDomains          []string             `mapstructure:"dns_domains,omitempty"`
	NameServers         []string             `mapstructure:"name_servers,omitempty,omitempty"`
	NtpServers          []string             `mapstructure:"ntp_servers,omitempty"`
	TimeZone            timeZone             `mapstructure:"timezone,omitempty"`
	ClusterCertificate  ClusterCertificate   `mapstructure:"certificate,omitempty"`
	ManagementInterface ClusterMgmtInterface `mapstructure:"management_interface,omitempty"`
	Password            string               `mapstructure:"password,omitempty"`
}

// ClusterLicense describes the License data model used in ClusterResourceBodyDataModelONTAP.
type ClusterLicense struct {
	Keys []string `mapstructure:"keys,omitempty"`
}

// ClusterMgmtInterface describes the Management Interface data model used in ClusterResourceBodyDataModelONTAP.
type ClusterMgmtInterface struct {
	IP ClusterMgmtInterfaceIP `mapstructure:"ip"`
}

// ClusterMgmtInterfaceIP describes the IP data model used in ClusterMgmtInterface.
type ClusterMgmtInterfaceIP struct {
	Address string `mapstructure:"address,omitempty"`
	Gateway string `mapstructure:"gateway,omitempty"`
	Netmask string `mapstructure:"netmask,omitempty"`
}

// timeZone describes the TimeZone data model used in ClusterGetDataModelONTAP.
type timeZone struct {
	Name string `mapstructure:"name,omitempty"`
}

// mgmtInterface describes the Management Interface data model used in ClusterGetDataModelONTAP.
type mgmtInterface struct {
	IP   ipAddress `mapstructure:"ip"`
	Name string    `mapstructure:"name"`
	ID   string    `mapstructure:"uuid"`
}

// ClusterCertificate describes the Certificate data model used in ClusterGetDataModelONTAP.
type ClusterCertificate struct {
	ID string `mapstructure:"uuid,omitempty"`
}

// versionModelONTAP describes the Version data model used in ClusterGetDataModelONTAP.
type versionModelONTAP struct {
	Full       string
	Generation int
	Major      int
	Minor      int
}

// ipAddress describes the IP data model used in mgmtInterface.
type ipAddress struct {
	Address string `mapstructure:"address"`
}

// noddMgmtInterface describes the Management Interface data model used in ClusterNodeGetDataModelONTAP.
type noddMgmtInterface struct {
	IP ipAddress `mapstructure:"ip"`
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
	query.Fields([]string{"name", "location", "contact", "dns_domains", "name_servers", "ntp_servers", "management_interfaces", "timezone", "certificate", "uuid"})
	// statusCode, response, err := r.GetNilOrOneRecord("cluster", query, nil)
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

// CreateCluster to create cluster. This is async operation.
func CreateCluster(errorHandler *utils.ErrorHandler, r restclient.RestClient, body ClusterResourceBodyDataModelONTAP) error {
	api := "cluster"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return errorHandler.MakeAndReportError("error encoding cluster body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	statusCode, response, err := r.CallCreateMethod(api, nil, bodyMap)
	if err != nil {
		return errorHandler.MakeAndReportError("error creating cluster", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create cluster source - udata: %#v", response))
	return nil
}

// UpdateCluster to update cluster. This is async operation.
func UpdateCluster(errorHandler *utils.ErrorHandler, r restclient.RestClient, body ClusterResourceBodyDataModelONTAP) error {
	api := "cluster"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return errorHandler.MakeAndReportError("error encoding cluster body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	statusCode, response, err := r.CallUpdateMethod(api, nil, bodyMap)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating cluster", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Update cluster source - udata: %#v", response))
	return nil
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

	var nodes []ClusterNodeGetDataModelONTAP
	for _, record := range records {
		var dataONTAP ClusterNodeGetDataModelONTAP
		if err := mapstructure.Decode(record, &dataONTAP); err != nil {
			return nil, errorHandler.MakeAndReportError("error decoding cluster nodes info", fmt.Sprintf("error: %s, statusCode %d, record %#v", err, statusCode, record))
		}
		nodes = append(nodes, dataONTAP)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read cluster data source NODES: %#v", nodes))
	return nodes, nil
}
