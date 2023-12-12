package interfaces

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// ClusterPeerGetDataModelONTAP describes the GET record data model using go types for mapping.
type ClusterPeerGetDataModelONTAP struct {
	Name             string                `mapstructure:"name"`
	UUID             string                `mapstructure:"uuid"`
	Remote           Remote                `mapstructure:"remote"`
	Status           Status                `mapstructure:"status"`
	PeerApplications []string              `mapstructure:"peer_applications"`
	Encryption       ClusterPeerEncryption `mapstructure:"encryption"`
	IPAddress        string                `mapstructure:"ip_address"`
	Ipspace          ClusterPeerIpspace    `mapstructure:"ipspace"`
}

// ClusterPeerIpspace describes the GET record data model using go types for mapping.
type ClusterPeerIpspace struct {
	Name string `mapstructure:"name"`
}

// ClusterPeerEncryption describes the GET record data model using go types for mapping.
type ClusterPeerEncryption struct {
	Propsed string `mapstructure:"proposed"`
	State   string `mapstructure:"state"`
}

// Remote describes the GET record data model using go types for mapping.
type Remote struct {
	IPAddress []string `mapstructure:"ip_addresses"`
	Name      string   `mapstructure:"name"`
}

// Status describes the GET record data model using go types for mapping.
type Status struct {
	State string `mapstructure:"state"`
}

// ClusterPeerResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type ClusterPeerResourceBodyDataModelONTAP struct {
	Name string `mapstructure:"name"`
	SVM  svm    `mapstructure:"svm"`
}

// ClusterPeerDataSourceFilterModel describes the data source data model for queries.
type ClusterPeerDataSourceFilterModel struct {
	Name string `mapstructure:"name"`
}

// GetClusterPeerByName gets a cluster peer by name.
func GetClusterPeerByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string) (*ClusterPeerGetDataModelONTAP, error) {
	query := r.NewQuery()
	query.Add("name", name)
	query.Fields([]string{"name", "uuid", "remote", "status", "peer_applications", "encryption", "ip_address", "ipspace"})
	statusCode, response, err := r.GetNilOrOneRecord("cluster/peers", query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("Error getting cluster peer", fmt.Sprint("error on get cluster/peer: #{err}"))
	}
	if response == nil {
		return nil, errorHandler.MakeAndReportError("No cluster peer found", fmt.Sprint("no cluster peer found with name: #{name}"))
	}
	var dataONTAP *ClusterPeerGetDataModelONTAP
	if error := mapstructure.Decode(response, &dataONTAP); error != nil {
		return nil, errorHandler.MakeAndReportError("Error decoding cluster peer", fmt.Sprintf("error decoding cluster peer: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read Cluster/peer source - udata: %#v", dataONTAP))
	return dataONTAP, nil
}

// GetClusterPeers gets all cluster peers.
func GetClusterPeers(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *ClusterPeerDataSourceFilterModel) ([]ClusterPeerGetDataModelONTAP, error) {
	query := r.NewQuery()
	query.Fields([]string{"name", "uuid", "remote", "status", "peer_applications", "encryption", "ip_address", "ipspace"})
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding cluster peer filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords("cluster/peers", query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("Error getting cluster peers", fmt.Sprint("error on get cluster/peers: #{err}"))
	}
	if response == nil {
		return nil, errorHandler.MakeAndReportError("No cluster peers found", fmt.Sprint("no cluster peers found with name: #{name}"))
	}
	var dataONTAP []ClusterPeerGetDataModelONTAP
	for _, info := range response {
		var record ClusterPeerGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError("Error decoding cluster peers", fmt.Sprintf("error decoding cluster peers: %s, statusCode %d, response %#v", err, statusCode, response))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read Cluster/peers source - udata: %#v", dataONTAP))
	return dataONTAP, nil
}
