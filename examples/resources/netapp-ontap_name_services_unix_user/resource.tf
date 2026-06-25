# creating a UNIX user
resource "netapp-ontap_unix_user" "example_unix_user" {
  cx_profile_name = "hw-cluster"
  name = "unix_user1"
  svm_name = "svm1"
  full_name = "Unix User1"
  primary_gid = 99
  id = 99
}
