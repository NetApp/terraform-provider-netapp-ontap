data "netapp-ontap_network_ethernet_port" "physical" {
  # required to know which system to interface with
  cx_profile_name = "svacluster"
  name            = "e0a"
}

data "netapp-ontap_network_ethernet_port" "vlan" {
  # required to know which system to interface with
  cx_profile_name = "svacluster"
  name            = "e0a-100"
}
