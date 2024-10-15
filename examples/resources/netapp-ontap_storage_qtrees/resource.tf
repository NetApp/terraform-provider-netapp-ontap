resource "netapp-ontap_qtrees" "storage_qtrees" {
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


