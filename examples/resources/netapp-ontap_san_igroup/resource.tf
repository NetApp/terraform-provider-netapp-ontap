resource "netapp-ontap_san_igroup" "protocols_san_igroups" {
  # required to know which system to interface with
  cx_profile_name = "cluster2"
  name = "test1"
  svm = {
    name = "carchi-test"
  }
  os_type = "linux"
  comment = "test"
}
