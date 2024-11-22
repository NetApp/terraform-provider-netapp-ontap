resource "netapp-ontap_network_ip_interface" "with_node_port" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name            = "testme"
  svm_name        = "ansibleSVM"

  ip = {
    address = "10.10.10.10"
    netmask = 20
  }

  location = {
    home_port = "e0a-300"
    home_node = "netapp_single-01"
  }
}

resource "netapp-ontap_network_ip_interface" "with_broadcast_domain" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name            = "testme"
  svm_name        = "ansibleSVM"

  ip = {
    address = "10.10.10.20"
    netmask = 20
  }

  location = {
    broadcast_domain = {
      name = "tf_test_mgmt_svm02"
    }
  }
  service_policy = "default-management"
}
