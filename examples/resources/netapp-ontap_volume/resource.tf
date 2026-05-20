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

# creating a volume
resource "netapp-ontap_volume" "example2" {
  cx_profile_name = "cluster4"
  name = "tfvolume"
  svm_name = "vs0"
  aggregates = [
    {
      name = "sti248_vsim_ocvs076c_aggr1"
    },
    {
      name = "sti248_vsim_ocvs076d_aggr1"
    },
  ]
  space = {
    size = 800
    size_unit = "mb"
  }
  autosize = {
    minimum = 200
    maximum = 1000
    shrink_threshold = 10
    grow_threshold = 50
    mode = "grow"
    size_unit = "mb"
  }
  snapshot_locking_enabled = true
  style = "flexgroup"
  constituents_per_aggregate = 2
}
