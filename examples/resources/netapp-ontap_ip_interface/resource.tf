resource "netapp-ontap_ip_interface_resource" "ip_interface" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  name = "testme"
  svm_name = "ansibleSVM"
  ip = {
    address = "10.10.10.10"
    netmask = 18
    }
  location = {
    home_port = "e0d"
    home_node = "laurentn-test-create-1-01"
  }
}
