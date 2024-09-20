data "netapp-ontap_network_ip_routes" "network_ip_routes" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  gateway = "10.10.10.254"
  filter = {
    svm_name = "*a*"
    destination = {
      address = "0.0.0.0",
      netmask = "24",
    }
    gateway = "10.*"
  }
}
