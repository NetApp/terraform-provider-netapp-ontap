package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// SVMPeerResourceModel describes the resource data model.
type SVMPeerResourceModel struct {
	Name         string   `mapstructure:"name,omitempty"`
	Applications []string `mapstructure:"applications"`
	SVM          SVM      `mapstructure:"svm"`
	Peer         Peer     `mapstructure:"peer"`
}

// SVMPeerAcceptResourceModel describes the resource data model.
type SVMPeerAcceptResourceModel struct {
	State string `mapstructure:"state"`
}

// SVMPeerUpdateResourceModel describes the resource data model.
type SVMPeerUpdateResourceModel struct {
	State        string   `mapstructure:"state,omitempty"`
	Applications []string `mapstructure:"applications,omitempty"`
}

// SVMPeerDataSourceModel describes the data source model.
type SVMPeerDataSourceModel struct {
	Name         string   `mapstructure:"name,omitempty"`
	UUID         string   `mapstructure:"uuid"`
	Applications []string `mapstructure:"applications"`
	SVM          SVM      `mapstructure:"svm"`
	Peer         Peer     `mapstructure:"peer"`
	State        string   `mapstructure:"state"`
}

// SVMPeerDataSourceFilterModel describes the data source data model for queries.
type SVMPeerDataSourceFilterModel struct {
	SVM  SVM  `mapstructure:"svm"`
	Peer Peer `mapstructure:"peer"`
}

// SVM describes the resource data model.
type SVM struct {
	Name string `mapstructure:"name"`
}

// Peer describes the body data model using go types for mapping.
type Peer struct {
	Cluster Cluster `mapstructure:"cluster"`
	SVM     SVM     `mapstructure:"svm"`
}

// PeerData describes the body data model using go types for mapping.
type PeerData struct {
	Cluster *Cluster `mapstructure:"cluster"`
	SVM     *SVM     `mapstructure:"svm"`
}

// SVMPeersGetDataModelONTAP describes the GET record data model using go types for mapping.
type SVMPeersGetDataModelONTAP struct {
	Name  string `mapstructure:"name,omitempty"`
	UUID  string `mapstructure:"uuid"`
	State string `mapstructure:"state"`
}

// SVMPeersResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type SVMPeersResourceBodyDataModelONTAP struct {
	Name string `mapstructure:"name,omitempty"`
	SVM  svm    `mapstructure:"svm"`
}

// GetSVMPeer to get SVMPeer info by uuid
func GetSVMPeer(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) (*SVMPeerDataSourceModel, error) {
	statusCode, response, err := r.GetNilOrOneRecord("svm/peers/"+uuid, nil, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading svm peer info", fmt.Sprintf("error on GET svm/peers: %s, statusCode %d", err, statusCode))
	}

	var dataONTAP *SVMPeerDataSourceModel
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("failed to decode response from GET svm peer", fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read svm peer info: %#v", dataONTAP))
	return dataONTAP, nil
}

// GetSVMPeersBySVMNameAndPeerSvmName to get svm_peers info
func GetSVMPeersBySVMNameAndPeerSvmName(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmName string, PeerSvmName string) (*SVMPeerDataSourceModel, error) {
	api := "svm/peers"
	query := r.NewQuery()
	fields := []string{"svm", "peer", "name", "applications", "state"}
	query.Add("svm.name", svmName)
	query.Add("peer.svm.name", PeerSvmName)
	query.Fields(fields)
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading svm_peers info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP SVMPeerDataSourceModel
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read svm_peers data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetSvmPeersByName to get data source list svm info
func GetSvmPeersByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *SVMPeerDataSourceFilterModel) ([]SVMPeerDataSourceModel, error) {
	api := "svm/peers"
	query := r.NewQuery()
	fields := []string{"svm", "peer", "name", "applications", "state"}
	query.Fields(fields)

	if filter != nil {
		if filter.Peer != (Peer{}) {
			if filter.Peer.Cluster != (Cluster{}) {
				query.Set("peer.cluster.name", filter.Peer.Cluster.Name)
			}
			if filter.Peer.SVM != (SVM{}) {
				query.Set("peer.svm.name", filter.Peer.SVM.Name)
			}
		}
		if filter.SVM != (SVM{}) {
			query.Set("svm.name", filter.SVM.Name)
		}
	}

	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading svm peers info", fmt.Sprintf("error on GET svm/peers: %s, statusCode %d", err, statusCode))
	}

	var dataONTAP []SVMPeerDataSourceModel
	for _, info := range response {
		var record SVMPeerDataSourceModel
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError("failed to decode response from GET svm peers", fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
		}
		dataONTAP = append(dataONTAP, record)
	}

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read svm peers info: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateSVMPeers to create svm_peers
func CreateSVMPeers(errorHandler *utils.ErrorHandler, r restclient.RestClient, body SVMPeerResourceModel) (*SVMPeersGetDataModelONTAP, error) {
	api := "svm/peers"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding svm_peers body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	query.Add("return_timeout", "15")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating svm_peers", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP SVMPeersGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding svm_peers info", fmt.Sprintf("error on decode storage/svm_peers info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create svm_peers source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// UpdateSVMPeers updates Snapmirror
func UpdateSVMPeers(errorHandler *utils.ErrorHandler, r restclient.RestClient, data any, uuid string) error {
	api := "svm/peers/" + uuid
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return errorHandler.MakeAndReportError("error encoding svm_peers body", fmt.Sprintf("error on encoding svm/peers body: %s, body: %#v", err, data))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	query.Add("return_timeout", "15")
	// API has no option to return records
	statusCode, _, err := r.CallUpdateMethod(api, query, body)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating svm_peers", fmt.Sprintf("error on PATCH svm/peers: %s, statusCode %d", err, statusCode))
	}
	return nil
}

// DeleteSVMPeers to delete svm_peers
func DeleteSVMPeers(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := "svm/peers/" + uuid
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting svm_peers", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
