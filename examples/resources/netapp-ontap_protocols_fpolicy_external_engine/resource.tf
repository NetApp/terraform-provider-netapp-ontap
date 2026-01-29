resource "netapp-ontap_protocols_fpolicy_external_engine" "protocols_fpolicy_external_engine" {
  # Required fields
  cx_profile_name = "cluster4"
  name           = "test_engine"
  svm_name       = "svm1"
  port           = 9002
  primary_servers = [
    "10.132.145.20",
    "10.132.145.21"
  ]

  # Optional configuration
  secondary_servers = [
    "10.132.145.22"
  ]

  type = "asynchronous"
  format = "protobuf"

  keep_alive_interval     = "PT3M"
  session_timeout         = "PT15S"
  request_cancel_timeout  = "PT30S"
  request_abort_timeout   = "PT1M"
  status_request_interval = "PT15S"
  server_progress_timeout = "PT1M21S"

  max_server_requests    = 860
  max_connection_retries = 10

  ssl_option = "no_auth"
  
  # Certificate configuration (when using SSL)
  certificate = {
    name          = "fpolicy_cert1"
    serial_number = "12345678922"
    ca            = "fpolicy_test_ca"
  }

  buffer_size = {
    send_buffer = 1424000
    recv_buffer = 4494304
  }

  resiliency = {
    enabled            = true
    directory_path     = "/"
    retention_duration = "PT2M"
  }
}
