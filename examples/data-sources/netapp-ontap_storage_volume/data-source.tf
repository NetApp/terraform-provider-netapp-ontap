data "netapp-ontap_storage_volume" "storage_volume" {
  cx_profile_name = "cluster4"
  name = "svm4_root"
  svm_name = "svm4"
}
