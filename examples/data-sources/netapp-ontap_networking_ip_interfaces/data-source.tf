data "netapp-ontap_networking_ip_interfaces" "ip_interfaces" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    name = "lif*"
    svm_name = "*"
    scope = "svm"
  }
}
