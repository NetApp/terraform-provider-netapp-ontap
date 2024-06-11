resource "netapp-ontap_svm" "example" {
  cx_profile_name = "cluster2"
  name = "tfsvm"
  ipspace = "test"
  comment = "test"
  snapshot_policy = "default-1weekly"
  //subtype = "dp_destination"
  language = "en_us.utf_8"
  aggregates = ["aggr1", "test"]
  max_volumes = "200"
}
