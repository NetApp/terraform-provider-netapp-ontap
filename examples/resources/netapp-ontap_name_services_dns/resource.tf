resource "netapp-ontap_dns" "dns" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  svm_name = "svm5"
  name_servers = ["1.1.1.1", "2.2.2.2"]
  dns_domains = ["foo.bar.com", "boo.bar.com"]
  skip_config_validation = true
}
