resource "netapp-ontap_dns" "dns" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  svm_name = "vs0"
  name_servers = ["2.2.2.2", "2.2.2.3"]
  dns_domains = ["test.boo.com", "test.foo.com"]
  skip_config_validation = true
  dynamic_dns = {
    enabled              = true
    fqdn                 = "testSVM.foo.bar.com"
    time_to_live         = "P1D"
    skip_fqdn_validation = false
    use_secure           = false
  }
}
