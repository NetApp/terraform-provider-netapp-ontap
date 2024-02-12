package interfaces

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// SecurityAccountResourceBodyDataModelONTAP describes the resource data model using go types for mapping.
type SecurityAccountResourceBodyDataModelONTAP struct {
	Name                       string                   `mapstructure:"name"`
	Applications               []map[string]interface{} `mapstructure:"applications,omitempty"`
	Owner                      SecurityAccountOwner     `mapstructure:"owner,omitempty"`
	Role                       SecurityAccountRole      `mapstructure:"role,omitempty"`
	Password                   string                   `mapstructure:"password,omitempty"`
	SecondAuthenticationMethod string                   `mapstructure:"second_authentication_method,omitempty"`
	Comment                    string                   `mapstructure:"comment,omitempty"`
	Locked                     bool                     `mapstructure:"locked,omitempty"`
}

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
	Name string `mapstructure:"name,omitempty"`
	UUID string `mapstructure:"uuid,omitempty"`
}

// SecurityAccountDataSourceFilterModel describes the data source filter data model.
type SecurityAccountDataSourceFilterModel struct {
	Name  string                `mapstructure:"name"`
	Owner *SecurityAccountOwner `mapstructure:"owner,omitempty"`
}

// GetSecurityAccountByName gets a security account by name.
func GetSecurityAccountByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, ownerName string) (*SecurityAccountGetDataModelONTAP, error) {
	query := r.NewQuery()
	query.Fields([]string{"name", "owner", "locked", "comment", "role", "scope", "applications"})
	query.Set("name", name)
	var statusCode int
	var response map[string]interface{}
	var err error
	if ownerName != "" {
		statusCode, response, err = r.GetNilOrOneRecord("security/accounts/"+ownerName+"/"+name, query, nil)
		if err != nil {
			return nil, errorHandler.MakeAndReportError("Error occurred when getting security account", fmt.Sprintf("error on get security/account: %s", err))
		}
	} else {
		statusCode, response, err = r.GetNilOrOneRecord("security/accounts", query, nil)
		if err != nil {
			return nil, errorHandler.MakeAndReportError("Error occurred when getting security account", fmt.Sprintf("error on get security/account: %s", err))
		}
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

// GetSecurityAccounts gets all security accounts.
func GetSecurityAccounts(errorHandler *utils.ErrorHandler, r restclient.RestClient, svnName string, name string) ([]SecurityAccountGetDataModelONTAP, error) {
	query := r.NewQuery()
	query.Fields([]string{"name", "owner", "locked", "comment", "role", "scope", "applications"})
	if svnName != "" {
		query.Set("owner.name", svnName)
	}
	if name != "" {
		query.Set("name", name)
	}

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("security account filter: %+v", query))
	statusCode, response, err := r.GetZeroOrMoreRecords("security/accounts", query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("Error occurred when getting security accounts", fmt.Sprintf("error on get security/accounts: %s", err))
	}
	if response == nil {
		return nil, errorHandler.MakeAndReportError("No Accounts found", fmt.Sprintf("No accounts found"))
	}
	var dataOntap []SecurityAccountGetDataModelONTAP
	for _, info := range response {
		var dataOntapItem SecurityAccountGetDataModelONTAP
		if error := mapstructure.Decode(info, &dataOntapItem); error != nil {
			return nil, errorHandler.MakeAndReportError("Error occurred when decoding security account", fmt.Sprintf("error on decoding security/account: %s, statusCode: %d, response %+v", error, statusCode, response))
		}
		dataOntap = append(dataOntap, dataOntapItem)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("security accounts: %+v", dataOntap))
	return dataOntap, nil
}

// CreateSecurityAccount creates a security account.
func CreateSecurityAccount(errorHandler *utils.ErrorHandler, r restclient.RestClient, body SecurityAccountResourceBodyDataModelONTAP) (*SecurityAccountGetDataModelONTAP, error) {
	api := "security/accounts"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("Error occurred when decoding security account", fmt.Sprintf("error on decoding security/account: %s", err))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("security account body: %+v", bodyMap))
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("Error occurred when creating security account", fmt.Sprintf("error on create security/account: %s, statusCode: %d, response %+v", err, statusCode, response))
	}
	var dataOntap SecurityAccountGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataOntap); err != nil {
		return nil, errorHandler.MakeAndReportError("Error occurred when decoding security account", fmt.Sprintf("error on decoding security/account: %s, statusCode: %d, response %+v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("security account: %+v", dataOntap))
	return &dataOntap, nil
}

// DeleteSecurityAccount deletes a security account.
func DeleteSecurityAccount(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, ownerID string) error {
	api := "security/accounts/" + ownerID + "/" + name
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("Error occurred when deleting security account", fmt.Sprintf("error on delete security/account: %s, statusCode: %d", err, statusCode))
	}
	return nil
}
