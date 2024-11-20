data "netapp-ontap_port" "physical" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name            = "e0a"
}

data "netapp-ontap_port" "lag" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name            = "a0a"
}

data "netapp-ontap_port" "vlan" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name            = "e0a-100"
}
