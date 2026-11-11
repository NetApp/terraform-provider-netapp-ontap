package protocols_test

import (
	"fmt"
	"os"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"	
)

func TestAccSVMAuditResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccSVMAuditResourceBasicConfig("tf_acc_svm"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "enabled", "false"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "log_path", "/"),
				),
			},
			// update and read
			// update log.retention
			{
				Config: testAccSVMAuditResourceUpdateLogRetentionConfig("tf_acc_svm"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "enabled", "false"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "log_path", "/"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "log.format", "xml"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "log.retention.count", "1"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "log.retention.duration", "PT0S"),
				),
			},
			// update and read
			// update log.rotation.schedule
			{
				Config: testAccSVMAuditResourceUpdateLogRotationConfig("tf_acc_svm"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "svm_name", "tf_acc_svm"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "enabled", "false"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "log_path", "/"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "log.rotation.schedule.days.#", "0"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "log.rotation.schedule.hours.#", "3"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "log.rotation.schedule.hours.0", "6"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "log.rotation.schedule.hours.1", "12"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "log.rotation.schedule.hours.2", "18"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "log.rotation.schedule.minutes.#", "3"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "log.rotation.schedule.minutes.0", "15"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "log.rotation.schedule.minutes.1", "30"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "log.rotation.schedule.minutes.2", "45"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "log.rotation.schedule.weekdays.#", "1"),
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "log.rotation.schedule.weekdays.0", "-1"),
				),
			},
			// Import and read
			{
				ResourceName:  "netapp-ontap_svm_audit.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s", "tf_acc_svm", "hw-cluster"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_svm_audit.example", "svm_name", "tf_acc_svm"),
				),
			},
		},
	})
}

func testAccSVMAuditResourceBasicConfig(svmName string) string {
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
	  name = "hw-cluster"
	  hostname = "%s"
	  username = "%s"
	  password = "%s"
	  validate_certs = false
	},
  ]
}

resource "netapp-ontap_svm_audit" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  enabled = false
  log_path = "/"
}`, host, admin, password, svmName)
}

func testAccSVMAuditResourceUpdateLogRetentionConfig(svmName string) string {
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
      name = "hw-cluster"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_svm_audit" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  enabled = false
  log_path = "/"
  log = {
    format = "xml"
    retention = {
      count = 1
    }
  }
}`, host, admin, password, svmName)
}

func testAccSVMAuditResourceUpdateLogRotationConfig(svmName string) string {
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
      name = "hw-cluster"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_svm_audit" "example" {
  # required to know which system to interface with
  cx_profile_name = "hw-cluster"
  svm_name = "%s"
  enabled = false
  log_path = "/"
  log = {
    format = "xml"
    rotation = {
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
        weekdays = [
          -1
        ]
	  }
    }
  }
  
}`, host, admin, password, svmName)
}
