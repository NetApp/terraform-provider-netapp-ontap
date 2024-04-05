package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNameServicesLDAPResource(t *testing.T) {
	svmName := "accsvm"
	credName := "cluster4"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// non-existant SVM return code 2621462. Must happen before create/read
			{
				Config:      testAccNameServicesLDAPResourceConfig("non-existant", "subtree"),
				ExpectError: regexp.MustCompile("2621462"),
			},
			// Test create
			{
				Config: testAccNameServicesLDAPResourceConfig("svm1", "subtree"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_name_services_ldap_resource.name_services_ldap", "svm_name", "svm1"),
					resource.TestCheckResourceAttr("netapp-ontap_name_services_ldap_resource.name_services_ldap", "servers.0", "1.1.1.1"),
				),
			},
			// Test update
			{
				Config: testAccNameServicesLDAPResourceConfig("svm1", "base"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_name_services_ldap_resource.name_services_ldap", "svm_name", "svm1"),
					resource.TestCheckResourceAttr("netapp-ontap_name_services_ldap_resource.name_services_ldap", "base_scope", "base"),
				),
			},
			// Test importing a resource
			{
				ResourceName:  "netapp-ontap_name_services_ldap_resource.name_services_ldap",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s", svmName, credName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_name_services_ldap_resource.name_services_ldap", "svm_name", svmName),
					resource.TestCheckResourceAttr("netapp-ontap_name_services_ldap_resource.name_services_ldap", "servers.0", "acc1.test.com"),
				),
			},
		},
	})
}

func testAccNameServicesLDAPResourceConfig(svmName string, baseScope string) string {
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

resource "netapp-ontap_name_services_ldap_resource" "name_services_ldap" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  svm_name = "%s"
  servers = ["1.1.1.1", "2.2.2.2"]
  base_scope = "%s"
  skip_config_validation = true
}
`, host, admin, password, svmName, baseScope)
}
