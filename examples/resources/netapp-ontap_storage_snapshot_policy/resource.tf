resource "netapp-ontap_storage_snapshot_policy_resource" "storage_snapshot_policy" {
  # required to know which system to interface with
  cx_profile_name = "cluster2"
  name = "testmesnapshotpolicy"
  svm_name = "tfSVM"
  comment = "This is a test for tf snapshot policy"
  enabled = false
  copies = [
    {
      count = 3
      schedule = {
        name = "daily"
      }
    },
    {
      count = 2
      schedule = {
        name = "hourly"
      }
    },
  ]
}