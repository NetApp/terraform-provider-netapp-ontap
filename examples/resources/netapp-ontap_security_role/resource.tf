resource "netapp-ontap_security_role" "security_role" {
  # required to know which system to interface with
  cx_profile_name = "cluster3"
  name = "testme"
  svm_name = "temp"
  privileges = [
  {
    access = "all"
    path = "lun"
  },
  {
	  access = "all"
	  path = "vserver"
	  query = "-vserver acc_test"
	}
  ]
}
