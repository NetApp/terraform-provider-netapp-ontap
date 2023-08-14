resource "netapp-ontap_svm_resource" "svm" {
  cx_profile_name = "cluster2"
  name = "tfsvm2"
}
resource "netapp-ontap_storage_volume_resource" "volume" {
  cx_profile_name = netapp-ontap_svm_resource.svm.cx_profile_name
  aggregates = ["aggr1"]
  name = "volume2"
  svm_name = netapp-ontap_svm_resource.svm.name
}
resource "netapp-ontap_storage_volume_snapshot_resource" "volume_snap" {
  cx_profile_name = netapp-ontap_svm_resource.svm.cx_profile_name
  name = "volume2_snap1"
  volume_uuid = netapp-ontap_storage_volume_resource.volume.uuid
}
