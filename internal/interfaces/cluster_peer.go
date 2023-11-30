package interfaces

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// TODO:
// copy this file to match you data source (should match internal/interfaces/cluster_peer.go)
// replace ClusterPeer with the name of the resource, following go conventions, eg IPInterface
// replace cluster_peer with the name of the resource, for logging purposes, eg ip_interface
// replace api_url with API, eg ip/interfaces
// delete these 5 lines

// ClusterPeerGetDataModelONTAP describes the GET record data model using go types for mapping.
type ClusterPeerGetDataModelONTAP struct {
	Name             string                `mapstructure:"name"`
	UUID             string                `mapstructure:"uuid"`
	Remote           Remote                `mapstructure:"remote"`
	Status           Status                `mapstructure:"status"`
	PeerApplications []string              `mapstructure:"peer_applications"`
	Encryption       ClusterPeerEncryption `mapstructure:"encryption"`
	IpAddress        string                `mapstructure:"ip_address"`
	Ipspace          ClusterPeerIpspace    `mapstructure:"ipspace"`
}

type ClusterPeerIpspace struct {
	Name string `mapstructure:"name"`
}

type ClusterPeerEncryption struct {
	Propsed string `mapstructure:"proposed"`
	State   string `mapstructure:"state"`
}

type Remote struct {
	IpAddress []string `mapstructure:"ip_addresses"`
	Name      string   `mapstructure:"name"`
}

type Status struct {
	State string `mapstructure:"state"`
}

// ClusterPeerResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type ClusterPeerResourceBodyDataModelONTAP struct {
	Name string `mapstructure:"name"`
	SVM  svm    `mapstructure:"svm"`
}

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
