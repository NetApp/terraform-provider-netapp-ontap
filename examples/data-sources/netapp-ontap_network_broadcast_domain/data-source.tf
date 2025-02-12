data "netapp-ontap_network_broadcast_domain" "example" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  ipspace         = "Default"
  name            = "Default"
}
