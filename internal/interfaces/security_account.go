package interfaces

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// TODO:
// copy this file to match you data source (should match internal/interfaces/security_account.go)
// replace SecurityAccount with the name of the resource, following go conventions, eg IPInterface
// replace security_account with the name of the resource, for logging purposes, eg ip_interface
// replace api_url with API, eg ip/interfaces
// delete these 5 lines

// SecurityAccountGetDataModelONTAP describes the GET record data model using go types for mapping.
type SecurityAccountGetDataModelONTAP struct {
	Name         string                       `mapstructure:"name"`
	Owner        SecurityAccountOwner         `mapstructure:"owner,omitempty"`
	Locked       bool                         `mapstructure:"locked,omitempty"`
	Comment      string                       `mapstructure:"comment,omitempty"`
	Role         SecurityAccountRole          `mapstructure:"role,omitempty"`
	Scope        string                       `mapstructure:"scope,omitempty"`
	Applications []SecurityAccountApplication `mapstructure:"applications,omitempty"`
}

// SecurityAccountApplication describes the application data model using go types for mapping.
type SecurityAccountApplication struct {
	Application                string   `mapstructure:"application,omitempty"`
	SecondAuthenticationMethod string   `mapstructure:"second_authentication_method,omitempty"`
	AuthenticationMethods      []string `mapstructure:"authentication_methods,omitempty"`
}

// SecurityAccountRole describes the role data model using go types for mapping.
type SecurityAccountRole struct {
	Name string `mapstructure:"name,omitempty"`
}

// SecurityAccountOwner describes the owner data model using go types for mapping.
type SecurityAccountOwner struct {
	Name string `mapstructure:"name"`
	UUID string `mapstructure:"uuid,omitempty"`
}

// SecurityAccountResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type SecurityAccountResourceBodyDataModelONTAP struct {
	Name string `mapstructure:"name"`
	SVM  svm    `mapstructure:"svm"`
}

func GetSecurityAccountByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, ownerName string) (*SecurityAccountGetDataModelONTAP, error) {
	query := r.NewQuery()
	query.Fields([]string{"name", "owner", "locked", "comment", "role", "scope", "applications"})
	statusCode, response, err := r.GetNilOrOneRecord("security/accounts/"+ownerName+"/"+name, query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("Error occurred when getting security account", fmt.Sprintf("error on get security/account: %s", err))
	}
	if response == nil {
		return nil, errorHandler.MakeAndReportError("No Account found", fmt.Sprintf("No account with name: %s", name))
	}
	var dataOntap *SecurityAccountGetDataModelONTAP
	if error := mapstructure.Decode(response, &dataOntap); error != nil {
		return nil, errorHandler.MakeAndReportError("Error occurred when decoding security account", fmt.Sprintf("error on decoding security/account: %s, statusCode: %d, response %+v", error, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("security account: %+v", dataOntap))
	return dataOntap, nil
}
