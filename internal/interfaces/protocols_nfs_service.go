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
	Enabled          bool      `mapstructure:"enabled"`
	Protocol         Protocol  `mapstructure:"protocol"`
	Root             Root      `mapstructure:"root"`
	Security         Security  `mapstructure:"security"`
	ShowmountEnabled bool      `mapstructure:"showmount_enabled"`
	Transport        Transport `mapstructure:"transport"`
	VstorageEnabled  bool      `mapstructure:"vstorage_enabled"`
	Windows          Windows   `mapstructure:"windows"`
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
	PermittedEncrptionTypes []string `mapstructure:"permitted_encryption_types"`
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
