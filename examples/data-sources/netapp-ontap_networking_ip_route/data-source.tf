data "netapp-ontap_networking_ip_route" "networking_ip_route" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  destination = {
    address = "0.0.0.0"
  }
  svm_name = "ansibleSVM"
  gateway = "10.193.176.1"
}
