data "netapp-ontap_ports" "example" {
  # required to know which system to interface with
  cx_profile_name = "svacluster"
  filter = {
    name  = "e0*"
    state = "up"
    type  = "vlan"
  }
}
