resource "netapp-ontap_cifs_local_user" "protocols_cifs_local_user" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "testme"
  svm_name = "ansibleSVM"
  password = "netapp123!"
}
