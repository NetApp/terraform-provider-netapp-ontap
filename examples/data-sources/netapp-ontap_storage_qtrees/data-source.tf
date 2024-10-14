data "netapp-ontap_qtrees" "storage_qtrees" {
  # required to know which system to interface with
  cx_profile_name = "cluster3"
  filter = {
    svm_name = "temp"
  }
}
