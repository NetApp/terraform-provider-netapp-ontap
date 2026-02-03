resource "netapp-ontap_s3_group" "s3_group" {
  cx_profile_name = "cluster5"
  name = "csahu_test_group"
  svm_name = "svm1"
  users = ["test_user1", "test_user2"]
  policies = ["test_policy1", "test_policy2"]
  comment = "test s3 group"
}