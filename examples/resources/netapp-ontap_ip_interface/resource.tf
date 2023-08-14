resource "netapp-ontap_networking_ip_interface_resource" "ip_interface" {
  # required to know which system to interface with
  cx_profile_name = "cluster5"
  name = "testme"
  svm_name = "automation"
  ip = {
    address = "10.10.10.10"
    netmask = 20
    }
  location = {
    home_port = "e0d"
    home_node = "swenjun-vsim2"
  }
}
