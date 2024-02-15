data "netapp-ontap_storage_flexcaches_data_source" "storage_flexcache" {
  # required to know which system to interface with
  cx_profile_name = "cluster5"
    filter = {
    name = "aggr*"
    svm_name = "automation"
  }
}
