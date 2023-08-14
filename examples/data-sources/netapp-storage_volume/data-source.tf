data "netapp-ontap_storage_volume_data_source" "example" {
  cx_profile_name = "cluster2"
  name = "terraformTest"
  svm_name = "ansibleSVM"
  //aggregates = ["aggr1"]
}