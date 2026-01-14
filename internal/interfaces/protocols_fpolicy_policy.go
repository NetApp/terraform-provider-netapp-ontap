package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// FpolicyPolicyEngine describes the engine model
type FpolicyPolicyEngine struct {
	Name string `mapstructure:"name"`
}

// FpolicyPolicyEvent describes the event model
type FpolicyPolicyEvent struct {
	Name string `mapstructure:"name"`
}

// FpolicyPolicyScope describes the scope model
type FpolicyPolicyScope struct {
	IncludeVolumes                  []string `mapstructure:"include_volumes,omitempty"`
	ExcludeVolumes                  []string `mapstructure:"exclude_volumes,omitempty"`
	IncludeShares                   []string `mapstructure:"include_shares,omitempty"`
	ExcludeShares                   []string `mapstructure:"exclude_shares,omitempty"`
	IncludeExtension                []string `mapstructure:"include_extension,omitempty"`
	ExcludeExtension                []string `mapstructure:"exclude_extension,omitempty"`
	IncludeExportPolicies           []string `mapstructure:"include_export_policies,omitempty"`
	ExcludeExportPolicies           []string `mapstructure:"exclude_export_policies,omitempty"`
	CheckExtensionsOnDirectories    bool     `mapstructure:"check_extensions_on_directories"`
	ObjectMonitoringWithNoExtension bool     `mapstructure:"object_monitoring_with_no_extension"`
}

// ProtocolsFpolicyPolicyGetDataModelONTAP describes the GET record data model using go types for mapping.
type ProtocolsFpolicyPolicyGetDataModelONTAP struct {
	Name                  string               `mapstructure:"name"`
	Engine                FpolicyPolicyEngine  `mapstructure:"engine"`
	Events                []FpolicyPolicyEvent `mapstructure:"events"`
	Priority              int64                `mapstructure:"priority"`
	Mandatory             bool                 `mapstructure:"mandatory"`
	Enabled               bool                 `mapstructure:"enabled"`
	AllowPrivilegedAccess bool                 `mapstructure:"allow_privileged_access,omitempty"`
	PrivilegedUser        string               `mapstructure:"privileged_user,omitempty"`
	PassthroughRead       bool                 `mapstructure:"passthrough_read"`
	Scope                 FpolicyPolicyScope   `mapstructure:"scope"`
	SVM                   SvmDataModelONTAP    `mapstructure:"svm"`
}

// ProtocolsFpolicyPolicyResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type ProtocolsFpolicyPolicyResourceBodyDataModelONTAP struct {
	Name                  string               `mapstructure:"name"`
	Engine                *FpolicyPolicyEngine `mapstructure:"engine,omitempty"`
	Events                []string             `mapstructure:"events"`
	Priority              int64                `mapstructure:"priority"`
	Mandatory             bool                 `mapstructure:"mandatory"`
	AllowPrivilegedAccess string               `mapstructure:"allow_privileged_access,omitempty"`
	PrivilegedUser        string               `mapstructure:"privileged_user,omitempty"`
	PassthroughRead       bool                 `mapstructure:"passthrough_read"`
	Scope                 *FpolicyPolicyScope  `mapstructure:"scope"`
}

// ProtocolsFpolicyPolicyResourceUpdateBodyDataModelONTAP describes the update body data model (for PATCH - without name field)
type ProtocolsFpolicyPolicyResourceUpdateBodyDataModelONTAP struct {
	Engine                *FpolicyPolicyEngine `mapstructure:"engine,omitempty"`
	Events                []string             `mapstructure:"events,omitempty"`
	Priority              *int64               `mapstructure:"priority,omitempty"`
	Mandatory             *bool                `mapstructure:"mandatory,omitempty"`
	AllowPrivilegedAccess *bool                `mapstructure:"allow_privileged_access,omitempty"`
	PrivilegedUser        string               `mapstructure:"privileged_user,omitempty"`
	PassthroughRead       *bool                `mapstructure:"passthrough_read,omitempty"`
	Scope                 *FpolicyPolicyScope  `mapstructure:"scope,omitempty"`
}

// ProtocolsFpolicyPolicyDataSourceFilterModel describes filter model
type ProtocolsFpolicyPolicyDataSourceFilterModel struct {
	Name    string `mapstructure:"name"`
	SVMName string `mapstructure:"svm.name"`
}

// GetProtocolsFpolicyPolicyByName to get protocols_fpolicy_policy info
func GetProtocolsFpolicyPolicyByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmUUID string) (*ProtocolsFpolicyPolicyGetDataModelONTAP, error) {
	api := fmt.Sprintf("protocols/fpolicy/%s/policies/%s", svmUUID, name)
	query := r.NewQuery()
	query.Set("return_timeout", "15")
	query.Fields([]string{"name", "engine", "events", "priority", "mandatory", "enabled", "allow_privileged_access", "privileged_user", "passthrough_read", "scope", "svm"})

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_fpolicy_policy info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ProtocolsFpolicyPolicyGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_fpolicy_policy source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetProtocolsFpolicyPolicies to get protocols_fpolicy_policy info for all resources matching a filter
func GetProtocolsFpolicyPolicies(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *ProtocolsFpolicyPolicyDataSourceFilterModel) ([]ProtocolsFpolicyPolicyGetDataModelONTAP, error) {
	api := "protocols/fpolicy/*/policies"
	query := r.NewQuery()
	query.Set("return_timeout", "15")
	query.Fields([]string{"name", "engine", "events", "priority", "mandatory", "enabled", "allow_privileged_access", "privileged_user", "passthrough_read", "scope", "svm"})

	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding protocols_fpolicy_policies filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}

	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_fpolicy_policy info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []ProtocolsFpolicyPolicyGetDataModelONTAP
	for _, info := range response {
		var record ProtocolsFpolicyPolicyGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_fpolicy_policies data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateProtocolsFpolicyPolicy to create protocols_fpolicy_policy
func CreateProtocolsFpolicyPolicy(errorHandler *utils.ErrorHandler, r restclient.RestClient, body ProtocolsFpolicyPolicyResourceBodyDataModelONTAP, svmUUID string) (*ProtocolsFpolicyPolicyGetDataModelONTAP, error) {
	api := fmt.Sprintf("protocols/fpolicy/%s/policies", svmUUID)
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding protocols_fpolicy_policy body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Set("return_records", "true")

	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating protocols_fpolicy_policy", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ProtocolsFpolicyPolicyGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding protocols_fpolicy_policy info", fmt.Sprintf("error on decode storage/protocols_fpolicy_policies info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create protocols_fpolicy_policy source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// UpdateProtocolsFpolicyPolicy to update protocols_fpolicy_policy
func UpdateProtocolsFpolicyPolicy(errorHandler *utils.ErrorHandler, r restclient.RestClient, body ProtocolsFpolicyPolicyResourceUpdateBodyDataModelONTAP, svmUUID string, name string) error {
	api := fmt.Sprintf("protocols/fpolicy/%s/policies/%s", svmUUID, name)
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return errorHandler.MakeAndReportError("error encoding protocols_fpolicy_policy body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Set("return_timeout", "15")

	statusCode, _, err := r.CallUpdateMethod(api, query, bodyMap)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating protocols_fpolicy_policy", fmt.Sprintf("error on PATCH %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}

// DeleteProtocolsFpolicyPolicy to delete protocols_fpolicy_policy
func DeleteProtocolsFpolicyPolicy(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmUUID string, name string) error {
	api := fmt.Sprintf("protocols/fpolicy/%s/policies/%s", svmUUID, name)
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting protocols_fpolicy_policy", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
