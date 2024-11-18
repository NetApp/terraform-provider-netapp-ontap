resource "netapp-ontap_port" "lag" {
  cx_profile_name     = "svacluster"
  broadcast_domain_id = "8c94e68d-d39c-11ed-90b5-00a0b8bcea9d"
  node_id             = "78b5bb16-d39c-11ed-868e-00505682eae9"
  type                = "lag"

  lag = {
    distribution_policy = "mac"
    member_ports_id = [
      "8934326a-d39c-11ed-90b5-00a0b8bcea9d",
      "89344e11-d39c-11ed-90b5-00a0b8bcea9d",
    ]
    mode = "singlemode"
  }
}

resource "netapp-ontap_port" "vlan" {
  cx_profile_name     = "svacluster"
  broadcast_domain_id = "a4de0311-8d2d-11ef-9ca8-00a0b8bc0407"
  node_id             = "78b5bb16-d39c-11ed-868e-00505682eae9"
  type                = "vlan"

  vlan = {
    base_port_id = "893366a0-d39c-11ed-90b5-00a0b8bcea9d"
    tag          = 300
  }
}
