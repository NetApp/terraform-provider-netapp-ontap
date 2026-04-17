resource "netapp-ontap_network_ipspace" "example" {
    cx_profile_name = "hw-cluster"
    name            = "csahu_ipspace1"
}
