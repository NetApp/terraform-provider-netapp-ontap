package protocols_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCifsLocalUserResourceAlias(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCifsLocalUserResourceConfigAliasMissingVars("non-existant"),
				ExpectError: regexp.MustCompile("Missing required argument"),
			},
			// create with basic argument
			// {
			// 	Config: testAccCifsLocalUserResourceConfigAlias("terraform", "user1"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		// check name
			// 		resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_local_user_resource.example1", "name", "user1"),
			// 		// check svm_name
			// 		resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_local_user_resource.example1", "svm_name", "terraform"),
			// 		// check ID
			// 		resource.TestCheckResourceAttrSet("netapp-ontap_protocols_cifs_local_user_resource.example1", "id"),
			// 	),
			// },
			// // update test
			// {
			// 	Config: testAccCifsLocalUserResourceConfigAlias("terraform", "newuser"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		// check renamed user name
			// 		resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_local_user_resource.example1", "name", "terraform"),
			// 		// check id
			// 		resource.TestCheckResourceAttrSet("netapp-ontap_protocols_cifs_local_user_resource.example1", "id")),
			// },
			// // Test importing a resource
			// {
			// 	ResourceName:  "netapp-ontap_protocols_cifs_local_user_resource.example1",
			// 	ImportState:   true,
			// 	ImportStateId: fmt.Sprintf("%s,%s,%s", "Administrator", "terraform", "cluster4"),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_local_user_resource.example1", "svm_name", "terraform"),
			// 		resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_local_user_resource.example1", "name", "Administrator"),
			// 		resource.TestMatchResourceAttr("netapp-ontap_protocols_nfs_export_policy_rule.example1", "description", regexp.MustCompile(`Built-in administrator account`)),
			// 		resource.TestCheckTypeSetElemAttr("netapp-ontap_protocols_nfs_export_policy_rule.example1", "membership.*", "Administrators"),
			// 		// check id
			// 		resource.TestCheckResourceAttrSet("netapp-ontap_protocols_cifs_local_user_resource.example1", "id"),
			// 	),
			// },
		},
	})
}

func testAccCifsLocalUserResourceConfigAliasMissingVars(svmName string) string {
	host := os.Getenv("TF_ACC_NETAPP_HOST5")
	admin := os.Getenv("TF_ACC_NETAPP_USER")
	password := os.Getenv("TF_ACC_NETAPP_PASS2")
	if host == "" || admin == "" || password == "" {
		fmt.Println("TF_ACC_NETAPP_HOST5, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS2 must be set for acceptance tests")
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

resource "netapp-ontap_protocols_cifs_local_user_resource" "example1" {
	cx_profile_name = "cluster4"
	svm_name = "%s"
}
`, host, admin, password, svmName)
}

// func testAccCifsLocalUserResourceConfigAlias(svmName string, userName string) string {
// 	host := os.Getenv("TF_ACC_NETAPP_HOST5")
// 	admin := os.Getenv("TF_ACC_NETAPP_USER")
// 	password := os.Getenv("TF_ACC_NETAPP_PASS2")
// 	if host == "" || admin == "" || password == "" {
// 		fmt.Println("TF_ACC_NETAPP_HOST5, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS2 must be set for acceptance tests")
// 		os.Exit(1)
// 	}
// 	return fmt.Sprintf(`
// provider "netapp-ontap" {
// 	connection_profiles = [
// 		{
// 			name = "cluster4"
// 			hostname = "%s"
// 			username = "%s"
// 			password = "%s"
// 			validate_certs = false
// 		},
// 	]
// }
// resource "netapp-ontap_protocols_cifs_local_user_resource" "example1" {
// 	cx_profile_name = "cluster4"
// 	svm_name = "%s"
// 	name = "%s"
// 	password = "password!!!"
// }
// `, host, admin, password, svmName, userName)
// }
