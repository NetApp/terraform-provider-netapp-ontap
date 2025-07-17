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
}


# creating a volume with snapshot_locking_enabled
resource "netapp-ontap_volume" "ontap_vol" {
  cx_profile_name = "cluster4"
  name = "terraformTest"
  svm_name = "ansibleSVM"
  aggregates = [
    {
      name = "aggr1"
    },
  ]
  space = {
    size = 20
    size_unit = "mb"
  }
  snapshot_locking_enabled = "true"
}