resource "netapp-ontap_storage_lun_resource" "storage_lun" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "lun2"
  svm_name = "carchi-test"
  volume_name = "lunTest"
  os_type = "linux"
  size = 1048576

}
