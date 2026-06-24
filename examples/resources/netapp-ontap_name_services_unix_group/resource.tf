resource "netapp-ontap_unix_group" "unix_group" {
  cx_profile_name = "hw-cluster"
  name = "unix_group"
  svm_name = "svm1"
  users = ["unix_user1", "unix_user2"]
  id = 99
}