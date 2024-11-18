data "netapp-ontap_ports" "physical" {
  # required to know which system to interface with
  cx_profile_name = "svacluster"
  filter = {
    name  = "e0*"
    state = "up"
    type  = "physical"
  }
}

data "netapp-ontap_ports" "lag" {
  # required to know which system to interface with
  cx_profile_name = "svacluster"
  filter = {
    name  = "a0*"
    state = "up"
    type  = "lag"
  }
}

data "netapp-ontap_ports" "vlan" {
  # required to know which system to interface with
  cx_profile_name = "svacluster"
  filter = {
    name  = "e0a-*"
    state = "up"
    type  = "vlan"
  }
}
