resource "netapp-ontap_san_lun-map" "protocols_san_lun-map" {
  # required to know which system to interface with
  cx_profile_name = "cluster2"
  svm = {
    name = "carchi-test"
  }
  lun = {
    name = "/vol/lunTest/ACC-import-lun"
  }
  igroup = {
    name = "test"
  }
  # logical_unit_number = 1
}
