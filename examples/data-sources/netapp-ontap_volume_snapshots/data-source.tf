data "netapp-ontap_volume_snapshots" "storage_volume_snapshots" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    name = "weekly.*"
    svm_name ="ansibleSVM"
    volume_name ="ansibleVolume12"
   }
}
