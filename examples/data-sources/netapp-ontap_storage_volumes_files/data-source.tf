data "netapp-ontap_storage_volumes_files_data_source" "storage_volumes_files" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  volume_name = "acc_test_peer_root"
  path = ".snapshot"
  svm_name = "acc_test_peer"
}
