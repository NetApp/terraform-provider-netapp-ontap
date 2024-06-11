resource "netapp-ontap_cifs_local_group" "protocols_cifs_local_group" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  svm_name = "svm3"
  name = "SERVER12\\testme"
}
