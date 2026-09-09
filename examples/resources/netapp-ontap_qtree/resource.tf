resource "netapp-ontap_qtree" "storage_qtree" {
  # required to know which system to interface with
  cx_profile_name = "cluster3"
  name = "qtree1"
  svm_name = "svm1"
  volume_name = "vol1"
  security_style = "unix" 
  user = {
    name = "nobody"
  }
  group = {
    name = "nobody"
  }
}