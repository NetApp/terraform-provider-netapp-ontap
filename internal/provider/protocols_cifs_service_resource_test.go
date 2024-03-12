package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCifsServiceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test error
			{
				Config:      testAccCifsServiceResourceConfigMissingVars("non-existant"),
				ExpectError: regexp.MustCompile("Missing required argument"),
			},
			// Create and read
			// {
			// 	Config: testAccCifsServiceResourceConfig("carchi-test", "false"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_service_resource.example", "svm_name", "carchi-test"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_service_resource.example", "is_active", "false"),
			// 	),
			// },
			// // update and read
			// {
			// 	Config: testAccCifsServiceResourceConfig("carchi-test", "true"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_service_resource.example", "svm_name", "carchi-test"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_service_resource.example", "is_active", "true"),
			// 	),
			// },
			// Import and read
			{
				ResourceName:  "netapp-ontap_protocols_cifs_service_resource.example",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "SVM3_SERVER", "svm3", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_service_resource.example", "name", "SVM3_SERVER"),
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_service_resource.example", "svm_name", "svm3"),
				),
			},
		},
	})
}

func testAccCifsServiceResourceConfigMissingVars(svmName string) string {
	return fmt.Sprintf(`
	resource "netapp-ontap_protocols_cifs_service_resource" "example1" {
		svm_name = "%s"
	}
	`, svmName)
}

func testAccCifsServiceResourceConfig(name string, svmName string) string {
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
resource "netapp-ontap_protocols_cifs_service_resource" "example" {
	cx_profile_name = "cluster4"
	svm_name = "%s"
	name = "%s"
	ad_domain = {
		fqdn = "example.com"
		user = "admin"
		password = "password"
	}
}
`, host, admin, password, svmName, name)
}
