resource "netapp-ontap_storage_flexcache" "storage_flexcache" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "acc_test_storage_flexcache_volume"
  svm_name = "acc_test"
  origins = [
    {
      volume = {
        name = "acc_test_storage_flexcache_origin_volume"
      },
      svm = {
        name = "acc_test"
      }
    }
  ]
  size = 400
  size_unit = "mb"
  guarantee = {
    type = "none"
  }
  dr_cache = false
  global_file_locking_enabled = false
  aggregates = [
    {
      name = "acc_test"
    }
  ]
}
