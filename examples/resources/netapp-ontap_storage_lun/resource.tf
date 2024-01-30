resource "netapp-ontap_storage_lun_resource" "storage_lun" {
  # required to know which system to interface with
  cx_profile_name = "cluster1"
  name = "testme"
}
