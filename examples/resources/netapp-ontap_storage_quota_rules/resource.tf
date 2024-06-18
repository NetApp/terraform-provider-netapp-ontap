resource "netapp-ontap_quota_rules" "storage_quota_rules" {
  # required to know which system to interface with
  cx_profile_name = "cluster2"
  volume = {
    name = "lunTest"
    }
  svm = {
    name = "carchi-test"
    }
  type = "tree"
  qtree = {
    name = ""
    }
  # users = [{
  #   name = ""
  #   }]
  # group = {
  #   name = ""
  #   }
  files = {
    hard_limit = 100
    soft_limit = 70
    }
}
