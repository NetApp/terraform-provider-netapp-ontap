data "netapp-ontap_ip_interfaces_data_source" "ip_interfaces" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  filter = {
    name = "laurent*"
  }
}

data "netapp-ontap_ip_interface_data_source" "ip_interface" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  name = "laurentn-vsim1_clus_1"
}
