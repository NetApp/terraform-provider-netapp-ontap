resource "netapp-ontap_storage_volume_snapshot_resource" "example" {
  cx_profile_name = "cluster1"
  name = "terraformTest12"
  vserver = "ansibleSVM"
  aggregates = ["aggr1"]
}
