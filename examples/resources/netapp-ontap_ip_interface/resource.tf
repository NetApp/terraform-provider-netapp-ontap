resource "netapp-ontap_networking_ip_interface_resource" "ip_interface" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "testme"
  svm_name = "ansibleSVM"
  ip = {
    address = "10.10.10.10"
    netmask = 20
    }
  location = {
    home_port = "e0c"
    home_node = "ontap_cluster_1-01"
  }
}
