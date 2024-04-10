resource "netapp-ontap_protocols_san_igroup_resource" "protocols_san_igroups" {
  # required to know which system to interface with
  cx_profile_name = "cluster2"
  name = "test1"
  svm = {
    name = "carchi-test"
  }
  os_type = "linux"
  comment = "test"
}
