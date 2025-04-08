# creating a volume
resource "netapp-ontap_volume" "example" {
  cx_profile_name = "cluster4"
  name = "terraformTest2"
  svm_name = "ansibleSVM"
  aggregates = [
    {
      name = "aggr2"
    },
  ]
  space = {
    size = 20
    size_unit = "mb"
  }
}

# creating a volume with autosize options
resource "netapp-ontap_volume" "example" {
  cx_profile_name = "cluster4"
  name = "terraformTest3"
  svm_name = "ansibleSVM"
  aggregates = [
    {
      name = "aggr2"
    },
  ]
  space = {
    size = 50
    size_unit = "mb"
  }
  autosize = {
    minimum = 20
    maximum = 60
    shrink_threshold = 10
    grow_threshold = 90
    mode = "off"
    size_unit = "mb"
  }
}
