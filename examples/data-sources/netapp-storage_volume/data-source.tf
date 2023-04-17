data "netapp-ontap_storage_volume_data_source" "example" {
  cx_profile_name = "cluster2"
  name = "terraformTest"
  vserver = "ansibleSVM"
  //aggregates = ["aggr1"]
}