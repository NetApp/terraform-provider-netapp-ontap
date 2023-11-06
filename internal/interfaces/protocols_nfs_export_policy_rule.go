package interfaces

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// ExportpolicyRuleResourceBodyDataModelONTAP describes the resource data model.
type ExportpolicyRuleResourceBodyDataModelONTAP struct {
	// SVM                 svm                 `mapstructure:"svm"`
	ClientsMatch        []map[string]string `mapstructure:"clients,omitempty"`
	RoRule              []string            `mapstructure:"ro_rule"`
	RwRule              []string            `mapstructure:"rw_rule"`
	Protocols           []string            `mapstructure:"protocols,omitempty"`
	AnonymousUser       string              `mapstructure:"anonymous_user,omitempty"`
	Superuser           []string            `mapstructure:"superuser,omitempty"`
	AllowDeviceCreation bool                `mapstructure:"allow_device_creation,omitempty"`
	NtfsUnixSecurity    string              `mapstructure:"ntfs_unix_security,omitempty"`
	ChownMode           string              `mapstructure:"chown_mode,omitempty"`
	AllowSuid           bool                `mapstructure:"allow_suid,omitempty"`
	Index               int64               `mapstructure:"index,omitempty"`
}

// ExportpolicyRuleResourceModel describes the resource data model.
type ExportpolicyRuleResourceModel struct {
	Svm                 SvmDataModelONTAP             `mapstructure:"svm,omitempty"`
	ClientsMatch        []map[string]string           `mapstructure:"clients,omitempty"`
	RoRule              []string                      `mapstructure:"ro_rule,omitempty"`
	RwRule              []string                      `mapstructure:"rw_rule,omitempty"`
	Protocols           []string                      `mapstructure:"protocols,omitempty"`
	AnonymousUser       string                        `mapstructure:"anonymous_user,omitempty"`
	Superuser           []string                      `mapstructure:"superuser,omitempty"`
	AllowDeviceCreation bool                          `mapstructure:"allow_device_creation,omitempty"`
	NtfsUnixSecurity    string                        `mapstructure:"ntfs_unix_security,omitempty"`
	ChownMode           string                        `mapstructure:"chown_mode,omitempty"`
	AllowSuid           bool                          `mapstructure:"allow_suid,omitempty"`
	Index               int64                         `mapstructure:"index"`
	ExportPolicy        ExportPolicyGetDataModelONTAP `mapstructure:"policy,omitempty"`
}

// ClientMatch describes the clients match struct
type ClientMatch struct {
	Match string `mapstructure:"match,omitempty"`
}

// ExportPolicyRuleGetDataModelONTAP describes the GET record data model using go types for mapping.
type ExportPolicyRuleGetDataModelONTAP struct {
	Svm                 SvmDataModelONTAP             `mapstructure:"svm,omitempty"`
	RoRule              []string                      `mapstructure:"ro_rule"`
	RwRule              []string                      `mapstructure:"rw_rule"`
	Protocols           []string                      `mapstructure:"protocols"`
	AnonymousUser       string                        `mapstructure:"anonymous_user"`
	Superuser           []string                      `mapstructure:"superuser"`
	AllowDeviceCreation bool                          `mapstructure:"allow_device_creation"`
	NtfsUnixSecurity    string                        `mapstructure:"ntfs_unix_security"`
	ChownMode           string                        `mapstructure:"chown_mode"`
	AllowSuid           bool                          `mapstructure:"allow_suid"`
	ClientsMatch        []ClientMatch                 `mapstructure:"clients"`
	Index               int64                         `mapstructure:"index"`
	ExportPolicy        ExportPolicyGetDataModelONTAP `mapstructure:"policy,omitempty"`
}

// ExportPolicyRuleDataSourceFilterModel describes filter model.
type ExportPolicyRuleDataSourceFilterModel struct {
	SVMName string `tfsdk:"svm_name"`
}

// CreateExportPolicyRule to create export policy rule
func CreateExportPolicyRule(errorHandler *utils.ErrorHandler, r restclient.RestClient, data ExportpolicyRuleResourceBodyDataModelONTAP, exportPolicyID string) (*ExportPolicyRuleGetDataModelONTAP, error) {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding export policy rule body", fmt.Sprintf("error on encoding export policy rule body: %s, body: %#v", err, data))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(fmt.Sprintf("protocols/nfs/export-policies/%s/rules", exportPolicyID), query, body)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating export policy rule", fmt.Sprintf("error on POST protocols/nfs/export-policies/%s/rules: %s, statusCode %d", exportPolicyID, err, statusCode))
	}

	var dataONTAP ExportPolicyRuleGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding export policies rule info", fmt.Sprintf("error on decode protocols/nfs/export-policies/%s/rules info: %s, statusCode %d, response %#v", exportPolicyID, err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create export policy source rule - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetExportPolicyRule to get export policy rule
func GetExportPolicyRule(errorHandler *utils.ErrorHandler, r restclient.RestClient, exportPolicyID string, index int64) (*ExportPolicyRuleGetDataModelONTAP, error) {
	api := "protocols/nfs/export-policies/" + exportPolicyID + "/rules/" + strconv.FormatInt(index, 10)
	statusCode, response, err := r.GetNilOrOneRecord(api, nil, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading export policy rule info", fmt.Sprintf("error on GET protocols/nfs/export-policies/%s/rules/%d: %s", exportPolicyID, index, err))
	}

	var dataONTAP ExportPolicyRuleGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding export policy rule info", fmt.Sprintf("error on decode %s: %s, statusCode %d, response %#v", api, err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read export policy rule source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetExportPolicyRuleSingle to get export policy rule.
func GetExportPolicyRuleSingle(errorHandler *utils.ErrorHandler, r restclient.RestClient, exportPolicyID string, index int64, version versionModelONTAP) (*ExportPolicyRuleGetDataModelONTAP, error) {
	query := r.NewQuery()
	fields := []string{"policy.name", "svm.name", "svm.uuid", "superuser", "protocols", "policy.name", "allow_device_creation",
		"chown_mode", "rw_rule", "index", "allow_suid", "ro_rule", "clients.match", "anonymous_user"}

	if version.Generation == 9 && version.Major > 10 {
		fields = append(fields, "ntfs_unix_security")
	}

	query.Fields(fields)

	statusCode, response, err := r.GetNilOrOneRecord("protocols/nfs/export-policies/"+exportPolicyID+"/rules/"+strconv.FormatInt(index, 10), query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading export policy rule info", fmt.Sprintf("error on GET protocols/nfs/export-policies/%s/rules/%d: %s", exportPolicyID, index, err))
	}

	var dataONTAP *ExportPolicyRuleGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding export policy rule info", fmt.Sprintf("error on decode protocols/nfs/export-policies/%s/rules/%d: %s, statusCode %d, response %#v", exportPolicyID, index, err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read export policy rule source - udata: %#v", dataONTAP))
	return dataONTAP, nil
}

// GetListExportPolicyRules to get protocols_nfs_export_policy_rules info
func GetListExportPolicyRules(errorHandler *utils.ErrorHandler, r restclient.RestClient, exportPolicyID string, filter *ExportPolicyRuleDataSourceFilterModel, version versionModelONTAP) ([]ExportPolicyRuleGetDataModelONTAP, error) {
	api := "protocols/nfs/export-policies/" + exportPolicyID + "/rules"
	query := r.NewQuery()

	if filter != nil {
		if filter.SVMName != "" {
			query.Add("svm.name", filter.SVMName)
		}
	}

	fields := []string{"policy.name", "svm.name", "svm.uuid", "superuser", "protocols", "policy.name", "allow_device_creation",
		"chown_mode", "rw_rule", "index", "allow_suid", "ro_rule", "clients.match", "anonymous_user"}

	if version.Generation == 9 && version.Major > 10 {
		fields = append(fields, "ntfs_unix_security")
	}
	query.Fields(fields)

	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_nfs_export_policy_rule info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []ExportPolicyRuleGetDataModelONTAP
	for _, info := range response {
		var record ExportPolicyRuleGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_nfs_export_policy_rule data source - udata: %#v", dataONTAP))
	return dataONTAP, nil
}

// UpdateExportPolicyRule to update export policy rule
func UpdateExportPolicyRule(errorHandler *utils.ErrorHandler, r restclient.RestClient, data ExportpolicyRuleResourceBodyDataModelONTAP, exportPolicyID string, index int64) (*ExportPolicyRuleGetDataModelONTAP, error) {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding export policy rule body", fmt.Sprintf("error on encoding export policy rule body: %s, body: %#v", err, data))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Update export policy source rule - body data: %#v", data))
	statusCode, response, err := r.CallUpdateMethod(fmt.Sprintf("protocols/nfs/export-policies/%s/rules/%d", exportPolicyID, index), nil, body)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error updating export policy rule", fmt.Sprintf("error on PATCH protocols/nfs/export-policies/%s/rules/%d: %s, statusCode %d", exportPolicyID, index, err, statusCode))
	}

	var dataONTAP ExportPolicyRuleGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding export policies rule info", fmt.Sprintf("error on decode protocols/nfs/export-policies/%s/rules/%d info: %s, statusCode %d, response %#v", exportPolicyID, index, err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Update export policy source rule - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteExportPolicyRule to delete export policy rule
func DeleteExportPolicyRule(errorHandler *utils.ErrorHandler, r restclient.RestClient, exportPolicyID string, index int64) error {
	statusCode, _, err := r.CallDeleteMethod("protocols/nfs/export-policies/"+exportPolicyID+"/rules/"+strconv.FormatInt(index, 10), nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting export policy rule", fmt.Sprintf("error on DELETE protocols/nfs/export-policies/%s/rules/%d: %s, statusCode %d", exportPolicyID, index, err, statusCode))
	}
	return nil
}
