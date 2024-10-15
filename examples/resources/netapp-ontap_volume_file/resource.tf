resource "netapp-ontap_volume_file" "volumes_file" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  path = "vol3"
  volume_name = "terraform"
  svm_name = "terraform"
  type = "directory"
  unix_permissions = "755"
}
