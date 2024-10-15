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
	Authentication   Authentication        `mapstructure:"authentication"`
	PeerApplications []string              `mapstructure:"peer_applications"`
	Encryption       ClusterPeerEncryption `mapstructure:"encryption"`
	IPAddress        string                `mapstructure:"ip_address"`
	Ipspace          ClusterPeerIpspace    `mapstructure:"ipspace"`
}

// ClusterPeersGetDataModelONTAP describes the GET record data model using go types for mapping.
type ClusterPeersGetDataModelONTAP struct {
	Name           string                       `mapstructure:"name"`
	UUID           string                       `mapstructure:"uuid"`
	Authentication AuthenticationCreateResponse `mapstructure:"authentication"`
}

// ClusterPeersResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type ClusterPeersResourceBodyDataModelONTAP struct {
	Name             string         `mapstructure:"name,omitempty"`
	Remote           RemoteBody     `mapstructure:"remote"`
	PeerApplications []string       `mapstructure:"peer_applications,omitempty"`
	Authentication   Authentication `mapstructure:"authentication"`
}

// ClusterPeerIpspace describes the GET record data model using go types for mapping.
type ClusterPeerIpspace struct {
	Name string `mapstructure:"name"`
}

// Authentication describes the POST record body model using go types for mapping.
type Authentication struct {
	State              string `mapstructure:"state,omitempty"`
	GeneratePassphrase bool   `mapstructure:"generate_passphrase,omitempty"`
	Passphrase         string `mapstructure:"passphrase,omitempty"`
}

// AuthenticationCreateResponse describes the POST record response model using go types for mapping.
type AuthenticationCreateResponse struct {
	Passphrase string `mapstructure:"passphrase,omitempty"`
}

// ClusterPeerEncryption describes the GET record data model using go types for mapping.
type ClusterPeerEncryption struct {
	Proposed string `mapstructure:"proposed"`
	State    string `mapstructure:"state"`
}

// Remote describes the GET record data model using go types for mapping.
type Remote struct {
	IPAddress []string `mapstructure:"ip_addresses"`
	Name      string   `mapstructure:"name"`
}

// RemoteBody describes the POST record body model using go types for mapping.
type RemoteBody struct {
	IPAddress []string `mapstructure:"ip_addresses"`
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
		return nil, errorHandler.MakeAndReportError("Error getting cluster peer", fmt.Sprintf("error on get cluster/peer: %s, statusCode %d", err, statusCode))
	}
	if response == nil {
		return nil, errorHandler.MakeAndReportError("No cluster peer found", fmt.Sprintf("no cluster peer found with name: %s, statusCode %d", name, statusCode))
	}
	var dataONTAP *ClusterPeerGetDataModelONTAP
	if error := mapstructure.Decode(response, &dataONTAP); error != nil {
		return nil, errorHandler.MakeAndReportError("Error decoding cluster peer", fmt.Sprintf("error decoding cluster peer: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read Cluster/peer source - udata: %#v", dataONTAP))
	return dataONTAP, nil
}

// GetClusterPeer to get ClusterPeer info by uuid
func GetClusterPeer(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) (*ClusterPeerGetDataModelONTAP, error) {
	statusCode, response, err := r.GetNilOrOneRecord("cluster/peers/"+uuid, nil, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading cluster peer info", fmt.Sprintf("error on GET cluster/peers: %s, statusCode %d", err, statusCode))
	}

	var dataONTAP *ClusterPeerGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("failed to decode response from GET cluster peer", fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read cluster peer info: %#v", dataONTAP))
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
		return nil, errorHandler.MakeAndReportError("Error getting cluster peers", fmt.Sprintf("error on get cluster/peers: %s, statusCode %d", err, statusCode))
	}
	if response == nil {
		return nil, errorHandler.MakeAndReportError("No cluster peers found", "no cluster peers fouund")
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

// CreateClusterPeers to create cluster_peers
func CreateClusterPeers(errorHandler *utils.ErrorHandler, r restclient.RestClient, body ClusterPeersResourceBodyDataModelONTAP) (*ClusterPeersGetDataModelONTAP, error) {
	api := "cluster/peers"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding cluster_peers body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	query.Add("return_timeout", "15")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating cluster_peers", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ClusterPeersGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding cluster_peers info", fmt.Sprintf("error on decode cluster/cluster_peers info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create cluster_peers source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// UpdateClusterPeers updates Cluster Peer
func UpdateClusterPeers(errorHandler *utils.ErrorHandler, r restclient.RestClient, data any, uuid string) error {
	api := "cluster/peers/" + uuid
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return errorHandler.MakeAndReportError("error encoding cluster_peers body", fmt.Sprintf("error on encoding cluster/peers body: %s, body: %#v", err, data))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	query.Add("return_timeout", "15")
	// API has no option to return records
	statusCode, _, err := r.CallUpdateMethod(api, query, body)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating cluster_peers", fmt.Sprintf("error on PATCH cluster/peers: %s, statusCode %d", err, statusCode))
	}
	return nil
}

// DeleteClusterPeers to delete cluster_peers
func DeleteClusterPeers(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := "cluster/peers"
	statusCode, _, err := r.CallDeleteMethod(api+"/"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting cluster_peers", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
