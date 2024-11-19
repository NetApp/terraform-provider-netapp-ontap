resource "netapp-ontap_port" "lag" {
  cx_profile_name = "svacluster"

  broadcast_domain = {
    ipspace = "Default"
    name    = "Default"
  }

  node = {
    name = "netapp_single-01"
  }

  type = "lag"

  lag = {
    distribution_policy = "mac"
    member_ports = [
      "e0b",
      "e0c",
    ]
    mode = "singlemode"
  }
}

resource "netapp-ontap_port" "vlan" {
  cx_profile_name = "svacluster"

  broadcast_domain = {
    ipspace = "tf_test"
    name    = "tf_test_data_svm02"
  }

  node = {
    name = "netapp_single-01"
  }

  type = "vlan"

  vlan = {
    base_port = "e0a"
    tag       = 300
  }
}
