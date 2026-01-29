package protocols_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccProtocolsFpolicyExternalEngineResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test error
			{
				Config:      testAccProtocolsFpolicyExternalEngineResourceConfig("non-existant", "test_engine_error", 9001),
				ExpectError: regexp.MustCompile("svm non-existant not found"),
			},
			// Create and read
			{
				Config: testAccProtocolsFpolicyExternalEngineResourceConfig("tf_acc_svm", "test_engine", 9001),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_protocols_fpolicy_external_engine.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_protocols_fpolicy_external_engine.example", "name", "test_engine"),
					resource.TestCheckResourceAttr("netapp-ontap_protocols_fpolicy_external_engine.example", "port", "9001"),
					resource.TestCheckResourceAttr("netapp-ontap_protocols_fpolicy_external_engine.example", "primary_servers.0", "10.10.10.20"),
				),
			},
			// Update and read
			{
				Config: testAccProtocolsFpolicyExternalEngineResourceConfigAdvanced("tf_acc_svm", "test_engine", 9002),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_protocols_fpolicy_external_engine.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_protocols_fpolicy_external_engine.example", "name", "test_engine"),
					resource.TestCheckResourceAttr("netapp-ontap_protocols_fpolicy_external_engine.example", "port", "9002"),
					resource.TestCheckResourceAttr("netapp-ontap_protocols_fpolicy_external_engine.example", "type", "asynchronous"),
					resource.TestCheckResourceAttr("netapp-ontap_protocols_fpolicy_external_engine.example", "format", "protobuf"),
					resource.TestCheckResourceAttr("netapp-ontap_protocols_fpolicy_external_engine.example", "max_server_requests", "500"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_protocols_fpolicy_external_engine.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "test_engine", "tf_acc_svm", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_protocols_fpolicy_external_engine.example", "name", "test_engine"),
					resource.TestCheckResourceAttr("netapp-ontap_protocols_fpolicy_external_engine.example", "svm_name", "tf_acc_svm"),
				),
			},
		},
	})
}

func testAccProtocolsFpolicyExternalEngineResourceConfig(svmName, engineName string, port int) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "cluster4"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_protocols_fpolicy_external_engine" "example" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  svm_name = "%s"
  name = "%s"
  port = %d
  primary_servers = ["10.10.10.20"]
}`, host, admin, password, svmName, engineName, port)
}

func testAccProtocolsFpolicyExternalEngineResourceConfigAdvanced(svmName, engineName string, port int) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS must be set for acceptance tests")
		os.Exit(1)
	}
	return fmt.Sprintf(`
provider "netapp-ontap" {
 connection_profiles = [
    {
      name = "cluster4"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_protocols_fpolicy_external_engine" "example" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  svm_name = "%s"
  name = "%s"
  port = %d
  primary_servers = ["10.10.10.20"]
  secondary_servers = ["10.10.10.21"]
  type = "asynchronous"
  format = "protobuf"
  ssl_option = "no_auth"
  max_server_requests = 500
  max_connection_retries = 5
  session_timeout = "PT30S"
  request_cancel_timeout = "PT20S"
  request_abort_timeout = "PT40S"
  keep_alive_interval = "PT5M"
  status_request_interval = "PT15S"
  server_progress_timeout = "PT1M"
  buffer_size = {
    send_buffer = 1024000
    recv_buffer = 2048000
  }
  resiliency = {
    enabled = true
    directory_path = "/"
    retention_duration = "PT4M"
  }
}`, host, admin, password, svmName, engineName, port)
}