data "netapp-ontap_san_igroups" "protocols_san_igroups" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  filter = {
    name = "igroup*"
  }
}
