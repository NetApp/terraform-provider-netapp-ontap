data "netapp-ontap_lun" "storage_lun" {
  # required to know which system to interface with
  cx_profile_name = "cluster2"
  name = "/vol/lunTest/ACC-import-lun"
  svm_name = "carchi-test"
  location = {
    volume = {
      name = "lunTest"
    }
  }
}
