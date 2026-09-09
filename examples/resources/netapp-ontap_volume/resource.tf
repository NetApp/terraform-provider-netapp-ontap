# creating a volume
resource "netapp-ontap_volume" "example1" {
  cx_profile_name = "cluster4"
  name = "terraformTest2"
  svm_name = "ansibleSVM"
  aggregates = [
    {
      name = "aggr1"
    },
  ]
  space = {
    size = 50
    size_unit = "mb"
  }
  autosize = {
    minimum = 40
    maximum = 60
    shrink_threshold = 10
    grow_threshold = 90
    mode = "grow"
    size_unit = "mb"
  }
  snapshot_locking_enabled = true
}

# restoring a volume from snapshot
resource "netapp-ontap_volume" "example2" {
  cx_profile_name = "cluster4"
  name = "terraformTest2"
  svm_name = "ansibleSVM"
  aggregates = [
    {
      name = "aggr1"
    },
  ]
  space = {
    size = 50
    size_unit = "mb"
  }
  autosize = {
    minimum = 40
    maximum = 60
    shrink_threshold = 10
    grow_threshold = 90
    mode = "grow"
    size_unit = "mb"
  }
  snapshot_locking_enabled = true
  restore_to = {
    snapshot = {
      name = "snap1"
    }
  }
}
