data "netapp-ontap_storage_lun" "storage_lun" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "/vol/ansibleVolume18/lun1"
  svm_name = "svm0"
  privileges = "test"
  location = {
    volume = {
      name = "ansibleVolume18"
    }
  }
}
