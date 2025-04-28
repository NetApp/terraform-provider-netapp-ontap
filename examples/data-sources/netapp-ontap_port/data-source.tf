data "netapp-ontap_port" "physical" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name            = "e0a"
  node = {
    name = "netapp_single-01"
  }
}

data "netapp-ontap_port" "lag" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name            = "a0a"
  node = {
    name = "netapp_single-01"
  }
}

data "netapp-ontap_port" "vlan" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name            = "e0a-101"
  node = {
    name = "netapp_single-01"
  }
}
