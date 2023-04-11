package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// TODO:
// copy this file to match you data source (should match internal/interfaces/name_services_dns.go)
// replace NameServicesDNS with the name of the resource, following go conventions, eg IPInterface
// replace name_services_dns with the name of the resource, for logging purposes, eg ip_interface
// replace api_url with API, eg ip/interfaces
// delete these 5 lines

// NameServicesDNSGetDataModelONTAP describes the GET record data model using go types for mapping.
type NameServicesDNSGetDataModelONTAP struct {
	Domains []string          `mapstructure:"domains"`
	Servers []string          `mapstructure:"servers"`
	SVM     SvmDataModelONTAP `mapstructure:"svm"`
}

// GetNameServicesDNS to get name_services_dns info
func GetNameServicesDNS(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmName string, scope string) (*NameServicesDNSGetDataModelONTAP, error) {
	api := "name-services/dns"
	query := r.NewQuery()
	if scope != "cluster" {
		query.Set("svm.name", svmName)
	}
	query.Set("scope", scope)
	query.Fields([]string{"domains", "servers"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading name_services_dns info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP NameServicesDNSGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read name_services_dns data source: %#v", dataONTAP))
	return &dataONTAP, nil
}
