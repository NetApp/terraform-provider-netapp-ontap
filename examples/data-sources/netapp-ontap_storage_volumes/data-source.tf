data "netapp-ontap_storage_volumes_data_source" "storage_volumes" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    svm_name = "svm*"
  }
}
