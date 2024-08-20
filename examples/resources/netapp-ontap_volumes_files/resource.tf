resource "netapp-ontap_volumes_files" "volumes_files" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  path = "vol3"
  volume_name = "terraform"
  svm_name = "terraform"
  type = "directory"
  unix_permissions = "755"
}
