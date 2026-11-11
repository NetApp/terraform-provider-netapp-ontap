data "netapp-ontap_network_ipspace" "ipspace" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  name            = "Default"
}
