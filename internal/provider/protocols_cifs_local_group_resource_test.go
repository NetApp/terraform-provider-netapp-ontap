package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCifsLocalGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCifsLocalGroupResourceConfigMissingVars("non-existant"),
				ExpectError: regexp.MustCompile("Missing required argument"),
			},
			// create with basic argument
			{
				Config: testAccCifsLocalGroupResourceConfig("ansibleSVM", "group1"),
				Check: resource.ComposeTestCheckFunc(
					// check name
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_local_group_resource.example1", "name", "group1"),
					// check svm_name
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_local_group_resource.example1", "svm_name", "ansibleSVM"),
					// check ID
					resource.TestCheckResourceAttrSet("netapp-ontap_protocols_cifs_local_group_resource.example1", "id"),
				),
			},
			// update test
			{
				Config: testAccCifsLocalGroupResourceConfig("ansibleSVM", "newgroup"),
				Check: resource.ComposeTestCheckFunc(
					// check renamed group name
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_local_group_resource.example1", "name", "newgroup"),
					// check id
					resource.TestCheckResourceAttrSet("netapp-ontap_protocols_cifs_local_group_resource.example1", "id")),
			},
			// Test importing a resource
			{
				ResourceName:  "netapp-ontap_protocols_cifs_local_group_resource.example1",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s", "Administrators", "ansibleSVM", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_local_group_resource.example1", "svm_name", "ansibleSVM"),
					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_local_group_resource.example1", "name", "Administrators"),
					resource.TestMatchResourceAttr("netapp-ontap_protocols_nfs_export_policy_rule_resource.example1", "description", regexp.MustCompile(`Built-in Administrators`)),
					// check id
					resource.TestCheckResourceAttrSet("netapp-ontap_protocols_cifs_local_group_resource.example1", "id"),
				),
			},
		},
	})
}

func testAccCifsLocalGroupResourceConfigMissingVars(svmName string) string {
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

resource "netapp-ontap_protocols_cifs_local_group_resource" "example1" {
	cx_profile_name = "cluster4"
	svm_name = "%s"
}
`, host, admin, password, svmName)
}

func testAccCifsLocalGroupResourceConfig(svmName string, groupName string) string {
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
resource "netapp-ontap_protocols_cifs_local_group_resource" "example1" {
	cx_profile_name = "cluster4"
	svm_name = "%s"
	name = "%s"
}
`, host, admin, password, svmName, groupName)
}
