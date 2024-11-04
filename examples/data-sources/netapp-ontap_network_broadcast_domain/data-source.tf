data "netapp-ontap_network_broadcast_domain" "example" {
  # required to know which system to interface with
  cx_profile_name = "svacluster"
  ipspace         = "Default"
  name            = "Default"
}
