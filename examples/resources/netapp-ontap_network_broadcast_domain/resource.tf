resource "netapp-ontap_network_broadcast_domain" "example" {
  # required to know which system to interface with
  cx_profile_name = "svacluster"
  ipspace         = "Default"
  mtu             = 1500
  name            = "testme"
}
