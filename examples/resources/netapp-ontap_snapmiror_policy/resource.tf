resource "netapp-ontap_snapmirror_policy_resource" "snapmirror_policy" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "carchitestme"
  svm_name = "ansibleSVM"
  identity_preservation = "full"
  comment = "comment1"
  type = "async"
}
