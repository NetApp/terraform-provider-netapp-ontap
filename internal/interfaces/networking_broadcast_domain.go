package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// BroadcastDomainGetDataModelONTAP describes the GET record data model using go types for mapping.
type BroadcastDomainGetDataModelONTAP struct {
	IPspace BroadcastDomainIPSpace `mapstructure:"ipspace"`
	MTU     int64                  `mapstructure:"mtu"`
	Name    string                 `mapstructure:"name"`
	Ports   []BroadcastDomainPort  `mapstructure:"ports"`
	UUID    string                 `mapstructure:"uuid"`
}

// BroadcastDomainResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type BroadcastDomainResourceBodyDataModelONTAP struct {
	IPspace BroadcastDomainIPSpace `mapstructure:"ipspace"`
	MTU     int64                  `mapstructure:"mtu"`
	Name    string                 `mapstructure:"name"`
	Ports   []BroadcastDomainPort  `mapstructure:"ports"`
	UUID    string                 `mapstructure:"uuid"`
}

// BroadcastDomainIPSpace describes an IP space specifically for broadcast domains.
type BroadcastDomainIPSpace struct {
	Name string `mapstructure:"name"`
	UUID string `mapstructure:"uuid"`
}

// BroadcastDomainPort describes an ethernet port specifically for broadcast domains.
type BroadcastDomainPort struct {
	Name string `mapstructure:"name"`
	UUID string `mapstructure:"uuid"`
}

// BroadcastDomainDataSourceFilterModel describes filter model.
type BroadcastDomainDataSourceFilterModel struct {
	IPspace string `tfsdk:"ipspace"`
	Name    string `tfsdk:"name"`
}

// GetBroadcastDomain to get broadcast_domain info
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

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read broadcast_domain data source: %#v", dataONTAP))

	return &dataONTAP, nil
}

// GetBroadcastDomainByName to get broadcast_domain info
func GetBroadcastDomainByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, ipspace, name string) (*BroadcastDomainGetDataModelONTAP, error) {
	api := "/network/ethernet/broadcast-domains/"
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

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read broadcast_domain data source: %#v", dataONTAP))

	return &dataONTAP, nil
}

// GetListBroadcastDomains to get broadcast_domain info for all resources matching a filter
func GetListBroadcastDomains(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *BroadcastDomainDataSourceFilterModel) ([]BroadcastDomainGetDataModelONTAP, error) {
	api := "network/ethernet/broadcast-domains/"
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

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read broadcast_domain data source: %#v", dataONTAP))

	return dataONTAP, nil
}
