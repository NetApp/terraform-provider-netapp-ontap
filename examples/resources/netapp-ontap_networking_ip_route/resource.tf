resource "netapp-ontap_networking_ip_routes" "networking_ip_route" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  svm_name = "ansibleSVM"
  destination = {
    address = "10.10.10.10"
    netmask = 24
    }
  gateway = "10.10.10.1"
  metric = 35
}
