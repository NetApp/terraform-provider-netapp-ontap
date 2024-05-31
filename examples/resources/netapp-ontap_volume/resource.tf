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
