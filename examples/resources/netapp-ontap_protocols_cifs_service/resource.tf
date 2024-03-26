resource "netapp-ontap_protocols_cifs_service_resource" "protocols_cifs_service" {
  # required to know which system to interface with
  cx_profile_name = "cluster_cifs"
  name = "tftestcifs"
  svm_name = "testSVM"
  ad_domain = {
    fqdn = "MYTFDOMAIN.COM"
    organizational_unit = "CN=Computers"
    user = "administrator"
    password = "Ab0xB@wks!"
  }
  #enabled  = true
}
