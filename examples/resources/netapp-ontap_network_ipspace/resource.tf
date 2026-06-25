resource "netapp-ontap_network_ipspace" "example_ipspace" {
    cx_profile_name = "hw-cluster"
    name            = "csahu_ipspace1"
}
