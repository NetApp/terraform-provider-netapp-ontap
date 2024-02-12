data "netapp-ontap_protocols_san_igroup_data_source" "protocols_san_igroup" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  name = "testme"
}
