data "netapp-ontap_qtree" "storage_qtree" {
  # required to know which system to interface with
  cx_profile_name = "cluster3"
  name = "tree10"
  volume_name =  "temp_root"
  svm_name = "temp"
}
