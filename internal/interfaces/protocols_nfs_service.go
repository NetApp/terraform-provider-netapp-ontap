package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// ProtocolsNfsServiceGetDataModelONTAP describes the GET record data model using go types for mapping.
type ProtocolsNfsServiceGetDataModelONTAP struct {
	Enabled          bool              `mapstructure:"enabled"`
	Protocol         Protocol          `mapstructure:"protocol"`
	Root             Root              `mapstructure:"root"`
	Security         Security          `mapstructure:"security"`
	ShowmountEnabled bool              `mapstructure:"showmount_enabled"`
	Transport        Transport         `mapstructure:"transport"`
	VstorageEnabled  bool              `mapstructure:"vstorage_enabled"`
	Windows          Windows           `mapstructure:"windows"`
	SVM              SvmDataModelONTAP `mapstructure:"svm"`
}

// ProtocolsNfsServiceResourceDataModelONTAP describes the GET record data model using go types for mapping.
type ProtocolsNfsServiceResourceDataModelONTAP struct {
	Enabled          bool              `mapstructure:"enabled"`
	Protocol         Protocol          `mapstructure:"protocol"`
	Root             Root              `mapstructure:"root"`
	Security         Security          `mapstructure:"security"`
	ShowmountEnabled bool              `mapstructure:"showmount_enabled"`
	Transport        Transport         `mapstructure:"transport"`
	VstorageEnabled  bool              `mapstructure:"vstorage_enabled"`
	Windows          Windows           `mapstructure:"windows"`
	SVM              SvmDataModelONTAP `mapstructure:"svm"`
}

// Protocol describes the GET record data model using go types for mapping.
type Protocol struct {
	V3Enabled   bool        `mapstructure:"v3_enabled"`
	V4IdDomain  string      `mapstructure:"v4_id_domain"`
	V40Enabled  bool        `mapstructure:"v40_enabled"`
	V40Features V40Features `mapstructure:"v40_features"`
	V41Enabled  bool        `mapstructure:"v41_enabled"`
	V41Features V41Features `mapstructure:"v41_features"`
}

// V40Features describes the GET record data model using go types for mapping.
type V40Features struct {
	ACLEnabled             bool `mapstructure:"acl_enabled"`
	ReadDelegationEnabled  bool `mapstructure:"read_delegation_enabled"`
	WriteDelegationEnabled bool `mapstructure:"write_delegation_enabled"`
}

// V41Features describes the GET record data model using go types for mapping.
type V41Features struct {
	ACLEnabled             bool `mapstructure:"acl_enabled"`
	PnfsEnabled            bool `mapstructure:"pnfs_enabled"`
	ReadDelegationEnabled  bool `mapstructure:"read_delegation_enabled"`
	WriteDelegationEnabled bool `mapstructure:"write_delegation_enabled"`
}

// Root describes the GET record data model using go types for mapping.
type Root struct {
	IgnoreNtACL              bool `mapstructure:"ignore_nt_acl"`
	SkipWritePermissionCheck bool `mapstructure:"skip_write_permission_check"`
}

// Security describes the GET record data model using go types for mapping.
type Security struct {
	ChownMode               string   `mapstructure:"chown_mode"`
	NtACLDisplayPermission  bool     `mapstructure:"nt_acl_display_permission"`
	NtfsUnixSecurity        string   `mapstructure:"ntfs_unix_security"`
	PermittedEncrptionTypes []string `mapstructure:"permitted_encryption_types,omitempty"`
	RpcsecContextIdel       int64    `mapstructure:"rpcsec_context_idle"`
}

// Transport describes the GET record data model using go types for mapping.
type Transport struct {
	TCP            bool  `mapstructure:"tcp_enabled"`
	TCPMaxXferSize int64 `mapstructure:"tcp_max_transfer_size"`
	UDP            bool  `mapstructure:"udp_enabled"`
}

// Windows describes the GET record data model using go types for mapping.
type Windows struct {
	DefaultUser                string `mapstructure:"default_user"`
	MapUnknownUIDToDefaultUser bool   `mapstructure:"map_unknown_uid_to_default_user"`
	V3MsDosClientEnabled       bool   `mapstructure:"v3_ms_dos_client_enabled"`
}

// NfsServicesFilterModel describes filter model
type NfsServicesFilterModel struct {
	SVMName string `mapstructure:"svm.name"`
}

// GetProtocolsNfsService to get protcols_nfs_service info
func GetProtocolsNfsService(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmName string, version versionModelONTAP) (*ProtocolsNfsServiceGetDataModelONTAP, error) {
	api := "protocols/nfs/services"
	query := r.NewQuery()
	query.Set("svm.name", svmName)
	var fields = []string{"svm.name", "svm.uuid", "protocol.v3_enabled", "protocol.v40_enabled", "protocol.v41_enabled",
		"protocol.v41_features.pnfs_enabled", "vstorage_enabled", "protocol.v4_id_domain", "transport.tcp_enabled",
		"transport.udp_enabled", "protocol.v40_features.acl_enabled", "protocol.v40_features.read_delegation_enabled",
		"protocol.v40_features.write_delegation_enabled", "protocol.v41_features.acl_enabled", "protocol.v41_features.read_delegation_enabled",
		"protocol.v41_features.write_delegation_enabled", "enabled"}
	if version.Generation == 9 && version.Major > 10 {
		fields = append(fields, "root.ignore_nt_acl", "root.skip_write_permission_check",
			"security.chown_mode", "security.nt_acl_display_permission", "security.ntfs_unix_security", "security.rpcsec_context_idle",
			"windows.default_user", "windows.map_unknown_uid_to_default_user", "windows.v3_ms_dos_client_enabled", "transport.tcp_max_transfer_size")
	}
	query.Fields(fields)

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protcols_nfs_service info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ProtocolsNfsServiceGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protcols_nfs_service data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetProtocolsNfsServices to get protocols_nfs_services info
func GetProtocolsNfsServices(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *NfsServicesFilterModel, version versionModelONTAP) ([]ProtocolsNfsServiceGetDataModelONTAP, error) {
	api := "protocols/nfs/services"
	query := r.NewQuery()

	fields := []string{"svm.name", "svm.uuid", "protocol.v3_enabled", "protocol.v40_enabled", "protocol.v41_enabled",
		"protocol.v41_features.pnfs_enabled", "vstorage_enabled", "protocol.v4_id_domain", "transport.tcp_enabled",
		"transport.udp_enabled", "protocol.v40_features.acl_enabled", "protocol.v40_features.read_delegation_enabled",
		"protocol.v40_features.write_delegation_enabled", "protocol.v41_features.acl_enabled", "protocol.v41_features.read_delegation_enabled",
		"protocol.v41_features.write_delegation_enabled", "enabled"}
	if version.Generation == 9 && version.Major > 10 {
		fields = append(fields, "root.ignore_nt_acl", "root.skip_write_permission_check",
			"security.chown_mode", "security.nt_acl_display_permission", "security.ntfs_unix_security", "security.rpcsec_context_idle",
			"windows.default_user", "windows.map_unknown_uid_to_default_user", "windows.v3_ms_dos_client_enabled", "transport.tcp_max_transfer_size")
	}
	query.Fields(fields)
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding protocols_nfs_service filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_nfs_service info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []ProtocolsNfsServiceGetDataModelONTAP
	for _, info := range response {
		var record ProtocolsNfsServiceGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protcols_nfs_service data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateProtocolsNfsService Create a NFS Service
func CreateProtocolsNfsService(errorHandler *utils.ErrorHandler, r restclient.RestClient, data ProtocolsNfsServiceResourceDataModelONTAP) (*ProtocolsNfsServiceGetDataModelONTAP, error) {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding NFS Service body", fmt.Sprintf("error on encoding protocols/nfs/services body: %s, body: %#v", err, data))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod("protocols/nfs/services", query, body)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating NFS services", fmt.Sprintf("error on POST protocols/nfs/services: %s, statusCode %d", err, statusCode))
	}
	var dataONTAP ProtocolsNfsServiceGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding NFS Services info", fmt.Sprintf("error on decode protocols/nfs/services info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create volume source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteProtocolsNfsService Deletes a NFS Service
func DeleteProtocolsNfsService(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	statusCode, _, err := r.CallDeleteMethod("protocols/nfs/services/"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting NFS Service", fmt.Sprintf("error on DELETE protocols/nfs/services: %s, statusCode %d", err, statusCode))
	}
	return nil
}

// UpdateProtocolsNfsService Update a NFS service
func UpdateProtocolsNfsService(errorHandler *utils.ErrorHandler, r restclient.RestClient, request ProtocolsNfsServiceResourceDataModelONTAP, uuid string) error {
	var body map[string]interface{}
	if err := mapstructure.Decode(request, &body); err != nil {
		return errorHandler.MakeAndReportError("error encoding NFS Services body", fmt.Sprintf("error on encoding NFS Services body: %s, body: %#v", err, request))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, _, err := r.CallUpdateMethod("protocols/nfs/services/"+uuid, query, body)
	if err != nil {
		return errorHandler.MakeAndReportError("error modifying NFS Service", fmt.Sprintf("error on PATCH rotocols/nfs/services/s: %s, statusCode %d", err, statusCode))
	}
	return nil
}
