resource "netapp-ontap_storage_qtree" "storage_qtree" {
  # required to know which system to interface with
  cx_profile_name = "cluster3"
  name = "testme5"
  svm_name = "temp"
  volume_name = "temp_root"
  security_style = "unix" 
  user = {
    name = "nobody"
  }
  group = {
    name = "nobody"
  }
}


