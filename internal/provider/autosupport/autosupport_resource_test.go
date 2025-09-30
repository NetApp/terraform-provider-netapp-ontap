package autosupport_test

import (
	"fmt"
	"os"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAutoSupportResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test creating AutoSupport configuration
			{
				Config: testAccAutoSupportResourceConfigBasic("cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_autosupport.example", "enabled", "true"),
					resource.TestCheckResourceAttr("netapp-ontap_autosupport.example", "transport", "smtp"),
					resource.TestCheckResourceAttr("netapp-ontap_autosupport.example", "from", "ontap@example.com"),
					resource.TestCheckResourceAttr("netapp-ontap_autosupport.example", "contact_support", "true"),
					resource.TestCheckResourceAttr("netapp-ontap_autosupport.example", "is_minimal", "false"),
					resource.TestCheckResourceAttr("netapp-ontap_autosupport.example", "ondemand_enabled", "true"),
					resource.TestCheckResourceAttr("netapp-ontap_autosupport.example", "smtp_encryption", "none"),
				),
			},
		},
	})
}

func testAccAutoSupportResourceConfigInvalid(profileName string) string {
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
      name = "%s"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_autosupport" "example" {
	cx_profile_name = "%s"
	enabled = true
	transport = "smtp"
	to_addresses = ["admin@example.com"]
	from = "ontap@example.com"
	contact_support = true
	mail_hosts = ["mail.example.com"]
	is_minimal = false
	ondemand_enabled = true
	smtp_encryption = "none"
	proxy_url = "proxy.example.com"
	partner_addresses = ["partner@example.com"]
}`, profileName, host, admin, password, profileName)
}

func testAccAutoSupportResourceConfigBasic(profileName string) string {
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
      name = "%s"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_autosupport" "example" {
	cx_profile_name = "%s"
	enabled = true
	transport = "smtp"
	to_addresses = ["admin@example.com"]
	from = "ontap@example.com"
	contact_support = true
	mail_hosts = ["mail.example.com"]
	is_minimal = false
	ondemand_enabled = true
	smtp_encryption = "none"
	proxy_url = "proxy.example.com"
	partner_addresses = ["partner@example.com"]
}`, profileName, host, admin, password, profileName)
}

func testAccAutoSupportResourceConfigUpdated(profileName string) string {
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
      name = "%s"
      hostname = "%s"
      username = "%s"
      password = "%s"
      validate_certs = false
    },
  ]
}

resource "netapp-ontap_autosupport" "example" {
	cx_profile_name = "%s"
	enabled = true
	transport = "smtp"
	to_addresses = ["admin@example.com", "backup@example.com"]
	from = "updated@example.com"
	contact_support = false
	mail_hosts = ["mail1.example.com", "mail2.example.com"]
	is_minimal = true
	ondemand_enabled = false
	smtp_encryption = "start_tls"
	proxy_url = "newproxy.example.com"
	partner_addresses = ["partner1@example.com", "partner2@example.com"]
}`, profileName, host, admin, password, profileName)
}
