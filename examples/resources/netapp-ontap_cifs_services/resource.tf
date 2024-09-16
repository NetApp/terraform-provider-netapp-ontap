resource "netapp-ontap_cifs_services" "protocols_cifs_service" {
  # required to know which system to interface with
  cx_profile_name = "clustercifs"
  name = "tftestcifs"
  svm_name = "testSVM"
  ad_domain = {
    fqdn = "mytfdomain.com"
    organizational_unit = "CN=Computers"
    user = "cifstest"
    password = "xxxxxx"
  }
}
