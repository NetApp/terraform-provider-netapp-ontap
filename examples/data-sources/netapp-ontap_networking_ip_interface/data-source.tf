data "netapp-ontap_networking_ip_interface_data_source" "ip_interface" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "cluster_mgmt"
}
