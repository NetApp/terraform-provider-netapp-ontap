data "netapp-ontap_networking_ip_interfaces_data_source" "ip_interfaces" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    name = "*svm*"
  }
}