package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// EthernetPortGetDataModelONTAP describes the GET record data model using go types for mapping.
// https://docs.netapp.com/us-en/ontap-restapi/ontap/get-network-ethernet-ports-.html#response
type EthernetPortGetDataModelONTAP struct {
	BroadcastDomain EthernetPortBroadcastDomain `mapstructure:"broadcast_domain"`
	Enabled         bool                        `mapstructure:"enabled"`
	InterfaceCount  int64                       `mapstructure:"interface_count"`
	LAG             EthernetPortLAG             `mapstructure:"lag"`
	MACAddress      string                      `mapstructure:"mac_address"`
	MTU             int64                       `mapstructure:"mtu"`
	Name            string                      `mapstructure:"name"`
	Node            EthernetPortNode            `mapstructure:"node"`
	RDMAProtocols   []string                    `mapstructure:"rdma_protocols"`
	Reachability    string                      `mapstructure:"reachability"`
	Speed           int64                       `mapstructure:"speed"`
	State           string                      `mapstructure:"state"`
	Type            string                      `mapstructure:"type"`
	UUID            string                      `mapstructure:"uuid"`
	VLAN            EthernetPortVLAN            `mapstructure:"vlan"`

	// DiscoveredDevices         []EthernetPortDiscoveredDevice `mapstructure:"discovered_devices"`
	// Metric                    EthernetPortMetric             `mapstructure:"metric"`
	// ReachableBroadcastDomains []EthernetPortBroadcastDomain  `mapstructure:"reachable_broadcast_domains"`
	// Statistics                EthernetPortStatistics         `mapstructure:"statistics"`
}

// EthernetPortResourceBodyDataModelONTAP describes the body data model using go types for mapping.
// https://docs.netapp.com/us-en/ontap-restapi/ontap/post-network-ethernet-ports.html#request-body
type EthernetPortResourceBodyDataModelONTAP struct {
	BroadcastDomain EthernetPortBroadcastDomain `mapstructure:"broadcast_domain,omitempty"`
	Enabled         bool                        `mapstructure:"enabled"`
	LAG             EthernetPortLAG             `mapstructure:"lag,omitempty"`
	Node            EthernetPortNode            `mapstructure:"node,omitempty"`
	Type            string                      `mapstructure:"type,omitempty"`
	UUID            string                      `mapstructure:"uuid,omitempty"`
	VLAN            EthernetPortVLAN            `mapstructure:"vlan,omitempty"`
}

// EthernetPortBroadcastDomain describes a broadcast domain specifically for ethernet ports.
type EthernetPortBroadcastDomain struct {
	IPSpace EthernetPortIPSpace `mapstructure:"ipspace,omitempty"`
	Name    string              `mapstructure:"name,omitempty"`
	UUID    string              `mapstructure:"uuid,omitempty"`
}

// EthernetPortIPSpace describes an IP space specifically for ethernet ports.
type EthernetPortIPSpace struct {
	Name string `mapstructure:"name,omitempty"`
}

// EthernetPortLAG describes a link aggregation group (LAG) specifically for ethernet ports.
type EthernetPortLAG struct {
	ActivePorts        []EthernetPortLAGPort `mapstructure:"active_ports,omitempty"`
	DistributionPolicy string                `mapstructure:"distribution_policy,omitempty"`
	MemberPorts        []EthernetPortLAGPort `mapstructure:"member_ports,omitempty"`
	Mode               string                `mapstructure:"mode,omitempty"`
}

// EthernetPortLAGPort describes a port specifically for LAGs.
type EthernetPortLAGPort struct {
	Name string           `mapstructure:"name,omitempty"`
	Node EthernetPortNode `mapstructure:"node,omitempty"`
	UUID string           `mapstructure:"uuid,omitempty"`
}

// EthernetPortNode describes a node specifically for ethernet ports.
type EthernetPortNode struct {
	Name string `mapstructure:"name,omitempty"`
	UUID string `mapstructure:"uuid,omitempty"`
}

// EthernetPortVLAN describes a virtual local area network (VLAN) specifically for ethernet ports.
type EthernetPortVLAN struct {
	BasePort EthernetPortVLANBasePort `mapstructure:"base_port,omitempty"`
	Tag      int64                    `mapstructure:"tag,omitempty"`
}

// EthernetPortVLANBasePort describes a base port specifically for VLANs
type EthernetPortVLANBasePort struct {
	Name string           `mapstructure:"name,omitempty"`
	Node EthernetPortNode `mapstructure:"node,omitempty"`
	UUID string           `mapstructure:"uuid,omitempty"`
}

// EthernetPortsDataSourceFilterModel describes filter model.
type EthernetPortsDataSourceFilterModel struct {
	Name  string `tfsdk:"name"`
	State string `tfsdk:"state"`
	Type  string `tfsdk:"type"`
}

// Retrieve ethernet port details
// https://docs.netapp.com/us-en/ontap-restapi/ontap/get-network-ethernet-ports-.html
func GetEthernetPort(errorHandler *utils.ErrorHandler, r restclient.RestClient, id string) (*EthernetPortGetDataModelONTAP, error) {
	api := "/network/ethernet/ports/" + id
	query := r.NewQuery()
	query.Fields([]string{
		"broadcast_domain",
		"enabled",
		"interface_count",
		"lag",
		"mac_address",
		"mtu",
		"name",
		"node",
		"rdma_protocols",
		"reachability",
		"speed",
		"state",
		"type",
		"uuid",
		"vlan",
	})

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no ethernet port with id '%s' found", id)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"Error Reading Ethernet Port Info",
			fmt.Sprintf("Error on GET %s: %s, statusCode %d.", api, err, statusCode),
		)
	}

	var dataONTAP EthernetPortGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			fmt.Sprintf("Failed To Decode Response From GET %s", api),
			fmt.Sprintf("Error: %s, statusCode %d, response %#v.", err, statusCode, response),
		)
	}

	tflog.Debug(
		errorHandler.Ctx,
		fmt.Sprintf("Read ethernet port data source: %#v", dataONTAP),
	)

	return &dataONTAP, nil
}

// Retrieve ethernet port details
// https://docs.netapp.com/us-en/ontap-restapi/ontap/get-network-ethernet-ports.html
func GetEthernetPortByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string) (*EthernetPortGetDataModelONTAP, error) {
	api := "/network/ethernet/ports/"
	query := r.NewQuery()
	query.Set("name", name)
	query.Fields([]string{
		"broadcast_domain",
		"enabled",
		"interface_count",
		"lag",
		"mac_address",
		"mtu",
		"name",
		"node",
		"rdma_protocols",
		"reachability",
		"speed",
		"state",
		"type",
		"uuid",
		"vlan",
	})

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no ethernet port with name '%s' found", name)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"Error Reading Ethernet Port Info",
			fmt.Sprintf("Error on GET %s: %s, statusCode %d.", api, err, statusCode),
		)
	}

	var dataONTAP EthernetPortGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			fmt.Sprintf("Failed To Decode Response From GET %s", api),
			fmt.Sprintf("Error: %s, statusCode %d, response %#v.", err, statusCode, response),
		)
	}

	tflog.Debug(
		errorHandler.Ctx,
		fmt.Sprintf("Read ethernet port data source: %#v", dataONTAP),
	)

	return &dataONTAP, nil
}

// Retrieve ethernet ports for the entire cluster
// https://docs.netapp.com/us-en/ontap-restapi/ontap/get-network-ethernet-ports.html
func GetListEthernetPorts(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *EthernetPortsDataSourceFilterModel) ([]EthernetPortGetDataModelONTAP, error) {
	api := "/network/ethernet/ports/"
	query := r.NewQuery()
	query.Fields([]string{
		"broadcast_domain",
		"enabled",
		"interface_count",
		"lag",
		"mac_address",
		"mtu",
		"name",
		"node",
		"rdma_protocols",
		"reachability",
		"speed",
		"state",
		"type",
		"uuid",
		"vlan",
	})

	if filter != nil {
		if filter.Name != "" {
			query.Set("name", filter.Name)
		}
		if filter.State != "" {
			query.Set("state", filter.State)
		}
		if filter.Type != "" {
			query.Set("type", filter.Type)
		}
	}

	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf(
			"no ethernet ports with name '%s', state '%s', and type '%s' found",
			filter.Name,
			filter.State,
			filter.Type,
		)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"Error Reading Ethernet Ports Info",
			fmt.Sprintf("Error on GET %s: %s, statusCode %d.", api, err, statusCode),
		)
	}

	var dataONTAP []EthernetPortGetDataModelONTAP
	for _, info := range response {
		var record EthernetPortGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(
				fmt.Sprintf("Failed To Decode Response From GET %s", api),
				fmt.Sprintf("Error: %s, statusCode %d, info %#v.", err, statusCode, info),
			)
		}
		dataONTAP = append(dataONTAP, record)
	}

	tflog.Debug(
		errorHandler.Ctx,
		fmt.Sprintf("Read ethernet port data source: %#v", dataONTAP),
	)

	return dataONTAP, nil
}

// Create a new VLAN or LAG (ethernet port)
// https://docs.netapp.com/us-en/ontap-restapi/ontap/post-network-ethernet-ports.html
func CreateEthernetPort(errorHandler *utils.ErrorHandler, r restclient.RestClient, body EthernetPortResourceBodyDataModelONTAP) (*EthernetPortGetDataModelONTAP, error) {
	api := "/network/ethernet/ports"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"Error Encoding Ethernet Port Body",
			fmt.Sprintf("Error on encoding %s body: %s, body: %#v.", api, err, body),
		)
	}

	// Ensure that *either* VLAN *or* LAG attribute is present (but not both)
	switch body.Type {
	case "lag":
		delete(bodyMap, "vlan")

		// Fix "lag.member_ports" data type
		if lag, ok := bodyMap["lag"].(map[string]interface{}); ok {

			// Delete "member_ports" key and recreate with correct type
			delete(lag, "member_ports")
			lag["member_ports"] = make([]map[string]interface{}, len(body.LAG.MemberPorts))

			// Populate "member_ports"
			for i, port := range body.LAG.MemberPorts {

				// Decode port info
				var portBodyMap map[string]interface{}
				mapstructure.Decode(port, &portBodyMap)

				// Write decoded port info to array
				if array, ok := lag["member_ports"].([]map[string]interface{}); ok {
					array[i] = portBodyMap
				}
			}
		}
	case "vlan":
		delete(bodyMap, "lag")
	}

	tflog.Trace(
		errorHandler.Ctx,
		fmt.Sprintf("ethernet port bodyMap: %#v", bodyMap),
	)

	query := r.NewQuery()
	query.Add("return_records", "true")

	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"Error Creating Ethernet Port",
			fmt.Sprintf("Error on POST %s: %s, statusCode %d.", api, err, statusCode),
		)
	}

	var dataONTAP EthernetPortGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"Error Decoding Ethernet Port Info",
			fmt.Sprintf("Error on decode ethernet port info: %s, statusCode %d, response %#v.", err, statusCode, response),
		)
	}

	tflog.Debug(
		errorHandler.Ctx,
		fmt.Sprintf("Create ethernet port resource: %#v", dataONTAP),
	)

	return &dataONTAP, nil
}

// Update VLAN or LAG (ethernet port) properties
// https://docs.netapp.com/us-en/ontap-restapi/ontap/patch-network-ethernet-ports-.html
func UpdateEthernetPort(errorHandler *utils.ErrorHandler, r restclient.RestClient, body EthernetPortResourceBodyDataModelONTAP, id, t string) error {
	api := "/network/ethernet/ports/" + id
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return errorHandler.MakeAndReportError(
			"Error Encoding Ethernet Port Body",
			fmt.Sprintf("Error on encoding %s body: %s, body: %#v.", api, err, body),
		)
	}

	// Ensure that *either* VLAN *or* LAG attribute is present (but not both)
	switch t {
	case "lag":
		delete(bodyMap, "vlan")

		// Fix "lag.member_ports" data type
		if lag, ok := bodyMap["lag"].(map[string]interface{}); ok {

			// Delete "member_ports" key and recreate with correct type
			delete(lag, "member_ports")
			lag["member_ports"] = make([]map[string]interface{}, len(body.LAG.MemberPorts))

			// Populate "member_ports"
			for i, port := range body.LAG.MemberPorts {

				// Decode port info
				var portBodyMap map[string]interface{}
				mapstructure.Decode(port, &portBodyMap)

				// Write decoded port info to array
				if array, ok := lag["member_ports"].([]map[string]interface{}); ok {
					array[i] = portBodyMap
				}
			}
		}
	case "vlan":
		delete(bodyMap, "lag")
	}

	tflog.Trace(
		errorHandler.Ctx,
		fmt.Sprintf("ethernet port bodyMap: %#v", bodyMap),
	)

	statusCode, _, err := r.CallUpdateMethod(api, nil, bodyMap)
	if err != nil {
		return errorHandler.MakeAndReportError(
			"Error Updating Ethernet Port",
			fmt.Sprintf("Error on PATCH %s: %s, statusCode %d.", api, err, statusCode),
		)
	}

	tflog.Debug(
		errorHandler.Ctx,
		fmt.Sprintf("Update ethernet port resource: %#v", bodyMap),
	)

	return nil
}

// Delete a VLAN or LAG (ethernet port)
// https://docs.netapp.com/us-en/ontap-restapi/ontap/delete-network-ethernet-ports-.html
func DeleteEthernetPort(errorHandler *utils.ErrorHandler, r restclient.RestClient, id string) error {
	api := "/network/ethernet/ports/" + id

	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError(
			"Error Deleting Ethernet Port",
			fmt.Sprintf("Error on DELETE %s: %s, statusCode %d.", api, err, statusCode),
		)
	}

	tflog.Debug(
		errorHandler.Ctx,
		fmt.Sprintf("Delete ethernet port resource: %s", id),
	)

	return nil
}
