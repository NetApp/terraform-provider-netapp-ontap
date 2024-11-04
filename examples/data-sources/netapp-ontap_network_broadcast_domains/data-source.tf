data "netapp-ontap_network_broadcast_domains" "example" {
  # required to know which system to interface with
  cx_profile_name = "svacluster"
  filter = {
    ipspace = "Default"
    name    = "*"
  }
}
