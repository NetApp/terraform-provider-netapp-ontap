resource "netapp-ontap_lun" "storage_lun" {
  # required to know which system to interface with
  cx_profile_name = "cluster2"
  name = "/vol/lunTest/test10"
  logical_unit = "test10"
  svm_name = "carchi-test"
  volume_name = "lunTest"
  os_type = "linux"
  size = 1048576

}
