data "netapp-ontap_storage_flexcache" "storage_flexcache" {
  # required to know which system to interface with
  cx_profile_name = "cluster5"
  name = "fc5"
  svm_name = "automation"
}
