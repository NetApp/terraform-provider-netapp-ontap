resource "netapp-ontap_snapshot_policy" "storage_snapshot_policy" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "testmesnapshotpolicy"
  svm_name = "abc-test"
  comment = "This is a test for tf snapshot policy upate"
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