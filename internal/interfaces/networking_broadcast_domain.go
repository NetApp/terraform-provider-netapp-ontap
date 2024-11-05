package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// BroadcastDomainGetDataModelONTAP describes the GET record data model using go types for mapping.
// https://docs.netapp.com/us-en/ontap-restapi/ontap/get-network-ethernet-broadcast-domains-.html#response
type BroadcastDomainGetDataModelONTAP struct {
	IPspace BroadcastDomainIPSpace `mapstructure:"ipspace"`
	MTU     int64                  `mapstructure:"mtu"`
	Name    string                 `mapstructure:"name"`
	Ports   []BroadcastDomainPort  `mapstructure:"ports"`
	UUID    string                 `mapstructure:"uuid"`
}

// BroadcastDomainResourceBodyDataModelONTAP describes the body data model using go types for mapping.
// https://docs.netapp.com/us-en/ontap-restapi/ontap/post-network-ethernet-broadcast-domains.html#request-body
type BroadcastDomainResourceBodyDataModelONTAP struct {
	IPspace BroadcastDomainIPSpace `mapstructure:"ipspace,omitempty"`
	MTU     int64                  `mapstructure:"mtu,omitempty"`
	Name    string                 `mapstructure:"name,omitempty"`
	Ports   []BroadcastDomainPort  `mapstructure:"ports,omitempty"`
	UUID    string                 `mapstructure:"uuid,omitempty"`
}

// BroadcastDomainIPSpace describes an IP space specifically for broadcast domains.
type BroadcastDomainIPSpace struct {
	Name string `mapstructure:"name,omitempty"`
	// UUID string `mapstructure:"uuid,omitempty"`
}

// BroadcastDomainPort describes an ethernet port specifically for broadcast domains.
type BroadcastDomainPort struct {
	Name string `mapstructure:"name"`
	// UUID string `mapstructure:"uuid"`
}

// BroadcastDomainDataSourceFilterModel describes filter model.
type BroadcastDomainDataSourceFilterModel struct {
	IPspace string `tfsdk:"ipspace"`
	Name    string `tfsdk:"name"`
}

// Retrieve broadcast domain details
// https://docs.netapp.com/us-en/ontap-restapi/ontap/get-network-ethernet-broadcast-domains-.html
func GetBroadcastDomain(errorHandler *utils.ErrorHandler, r restclient.RestClient, id string) (*BroadcastDomainGetDataModelONTAP, error) {
	api := "/network/ethernet/broadcast-domains/" + id
	query := r.NewQuery()
	query.Fields([]string{
		"ipspace",
		"mtu",
		"name",
		"ports",
		"uuid",
	})

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no broadcast-domain with id '%s' found", id)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error reading broadcast-domain info",
			fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP BroadcastDomainGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response),
		)
	}

	tflog.Debug(
		errorHandler.Ctx,
		fmt.Sprintf("Read broadcast_domain data source: %#v", dataONTAP),
	)

	return &dataONTAP, nil
}

// Retrieve broadcast domain details
// https://docs.netapp.com/us-en/ontap-restapi/ontap/get-network-ethernet-broadcast-domains.html
func GetBroadcastDomainByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, ipspace, name string) (*BroadcastDomainGetDataModelONTAP, error) {
	api := "/network/ethernet/broadcast-domains"
	query := r.NewQuery()
	query.Set("ipspace", ipspace)
	query.Set("name", name)
	query.Fields([]string{
		"ipspace",
		"mtu",
		"name",
		"ports",
		"uuid",
	})

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no broadcast-domain with ipspace '%s' and name '%s' found", ipspace, name)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error reading broadcast-domain info",
			fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP BroadcastDomainGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response),
		)
	}

	tflog.Debug(
		errorHandler.Ctx,
		fmt.Sprintf("Read broadcast_domain data source: %#v", dataONTAP),
	)

	return &dataONTAP, nil
}

// Retrieve broadcast domains for the entire cluster
// https://docs.netapp.com/us-en/ontap-restapi/ontap/get-network-ethernet-broadcast-domains.html
func GetListBroadcastDomains(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *BroadcastDomainDataSourceFilterModel) ([]BroadcastDomainGetDataModelONTAP, error) {
	api := "network/ethernet/broadcast-domains"
	query := r.NewQuery()
	query.Fields([]string{
		"ipspace",
		"mtu",
		"name",
		"ports",
		"uuid",
	})

	if filter != nil {
		if filter.IPspace != "" {
			query.Set("ipspace", filter.IPspace)
		}
		if filter.Name != "" {
			query.Set("name", filter.Name)
		}
	}

	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no broadcast-domains with ipspace '%s' and name '%s' found", filter.IPspace, filter.Name)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error reading broadcast-domains info",
			fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP []BroadcastDomainGetDataModelONTAP
	for _, info := range response {
		var record BroadcastDomainGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(
				fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info),
			)
		}
		dataONTAP = append(dataONTAP, record)
	}

	tflog.Debug(
		errorHandler.Ctx,
		fmt.Sprintf("Read broadcast_domain data source: %#v", dataONTAP),
	)

	return dataONTAP, nil
}

// Create a new broadcast domain
// https://docs.netapp.com/us-en/ontap-restapi/ontap/post-network-ethernet-broadcast-domains.html
func CreateBroadcastDomain(errorHandler *utils.ErrorHandler, r restclient.RestClient, body BroadcastDomainResourceBodyDataModelONTAP) (*BroadcastDomainGetDataModelONTAP, error) {
	api := "/network/ethernet/broadcast-domains"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error encoding broadcast-domain body",
			fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body),
		)
	}
	query := r.NewQuery()
	query.Add("return_records", "true")

	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error creating broadcast-domain",
			fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP BroadcastDomainGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error decoding broadcast-domain info",
			fmt.Sprintf("error on decode broadcast-domain info: %s, statusCode %d, response %#v", err, statusCode, response),
		)
	}

	tflog.Debug(
		errorHandler.Ctx,
		fmt.Sprintf("Create broadcast_domain resource: %#v", dataONTAP),
	)

	return &dataONTAP, nil
}

// Update broadcast domain properties
// https://docs.netapp.com/us-en/ontap-restapi/ontap/patch-network-ethernet-broadcast-domains-.html
func UpdateBroadcastDomain(errorHandler *utils.ErrorHandler, r restclient.RestClient, body BroadcastDomainResourceBodyDataModelONTAP, id string) error {
	api := "/network/ethernet/broadcast-domains/" + id
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return errorHandler.MakeAndReportError(
			"error encoding broadcast-domain body",
			fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body),
		)
	}

	statusCode, _, err := r.CallUpdateMethod(api, nil, bodyMap)
	if err != nil {
		return errorHandler.MakeAndReportError(
			"error updating broadcast-domain",
			fmt.Sprintf("error on PATCH %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	tflog.Debug(
		errorHandler.Ctx,
		fmt.Sprintf("Update broadcast_domain resource: %#v", bodyMap),
	)

	return nil
}

// Delete a broadcast domain
// https://docs.netapp.com/us-en/ontap-restapi/ontap/delete-network-ethernet-broadcast-domains-.html
func DeleteBroadcastDomain(errorHandler *utils.ErrorHandler, r restclient.RestClient, id string) error {
	api := "/network/ethernet/broadcast-domains/" + id

	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError(
			"error deleting broadcast-domain",
			fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	tflog.Debug(
		errorHandler.Ctx,
		fmt.Sprintf("Delete broadcast_domain resource: %s", id),
	)

	return nil
}
