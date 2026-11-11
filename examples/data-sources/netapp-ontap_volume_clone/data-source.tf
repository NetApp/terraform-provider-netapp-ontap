data "netapp-ontap_volume_clone" "volume_clone" {
  cx_profile_name = "hw-cluster"
  name = "vol1_clone1"
  svm_name = "svm1"
}
