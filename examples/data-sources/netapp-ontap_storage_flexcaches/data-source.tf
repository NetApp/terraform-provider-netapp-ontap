data "netapp-ontap_storage_flexcaches_data_source" "storage_flexcache" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
    filter = {
    name = "f*"
    svm_name = "automation"
  }
}
