resource "netapp-ontap_storage_flexcache_resource" "storage_flexcache" {
  # required to know which system to interface with
  cx_profile_name = "cluster5"
  name = "fc10"
  svm_name = "automation"
  origins = [
    {
      volume = {
        name = "vol1"
      },
      svm = {
        name = "automation"
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
      name = "aggr1"
    }
  ]
}
