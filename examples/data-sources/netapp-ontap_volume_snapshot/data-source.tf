data "netapp-ontap_volume_snapshot" "snapshot" {
  cx_profile_name = "cluster4"
  name = "weekly.2023-10-08_0015"
  svm_name ="ansibleSVM"
  volume_name ="ansibleVolume12"
}
