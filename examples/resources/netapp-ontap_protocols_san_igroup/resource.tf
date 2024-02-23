resource "netapp-ontap_protocols_san_igroup_resource" "protocols_san_igroup" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  name = "testme"
}
