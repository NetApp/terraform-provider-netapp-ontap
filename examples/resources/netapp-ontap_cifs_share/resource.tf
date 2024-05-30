resource "netapp-ontap_cifs_share" "protocols_cifs_share" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "testme"
  path = "/wenjun_vol"
  svm_name = "ansibleSVM"
  acls = [
              {
                "permission": "read",
                "type": "windows",
                "user_or_group": "Everyone"
              }
            ]
  comment = "abedf"
}
