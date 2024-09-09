data "netapp-ontap_san_igroup" "protocols_san_igroup" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "igroup1"
  svm_name="svm0"
}
