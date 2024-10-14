data "netapp-ontap_network_ip_interface" "ip_interface" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "cluster_mgmt"
}
