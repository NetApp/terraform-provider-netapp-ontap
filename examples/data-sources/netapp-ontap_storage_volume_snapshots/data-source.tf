data "netapp-ontap_storage_volume_snapshots_data_source" "storage_volume_snapshots" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    name = "weekly.*"
    svm_name ="ansibleSVM"
    volume_name ="ansibleVolume12"
   }
}
