resource "netapp-ontap_svm_audit" "svm_audit_config" {
  cx_profile_name = "hw-cluster"
  svm_name = "svm1"
  enabled = false
  events = {
    authorization_policy = false
    cap_staging = false
    cifs_logon_logoff = true
    file_operations = true
    file_share = false
    security_group = false
    user_account = false
  }
  log_path = "/"
  log = {
    format = "xml"
    retention = {
      duration: "P1DT1H30M5S"
    }
    rotation ={
      /* The audit logs are rotated in January and March on Monday, Wednesday, and Friday,
      at 6:15, 6:30, 6:45, 12:15, 12:30, 12:45, 18:15, 18:30, and 18:45 */
      schedule = {
        hours = [
          6,
          12,
          18
        ]
        minutes = [
          15,
          30,
          45
        ]
        months = [
          1,
          3
        ]
        weekdays = [
          1,
          3,
          5
        ]
      }
    }
  }
}
