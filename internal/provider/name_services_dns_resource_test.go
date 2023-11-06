package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"os"
	"regexp"
	"testing"
)

func TestAccNameServicesDNSResource(t *testing.T) {
	svmName := "ansibleSVM"
	credName := "cluster4"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// non-existant SVM return code 2621462. Must happen before create/read
			{
				Config:      testAccNameServicesDNSResourceConfig("non-existant"),
				ExpectError: regexp.MustCompile("2621462"),
			},
			// Test importing a resource
			{
				ResourceName:  "netapp-ontap_name_services_dns_resource.name_services_dns",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s", svmName, credName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_name_services_dns_resource.name_services_dns", "svm_name", "ansibleSVM"),
					resource.TestCheckResourceAttr("netapp-ontap_name_services_dns_resource.name_services_dns", "name_servers.0", "netappad.com"),
				),
			},
		},
	})
}

func testAccNameServicesDNSResourceConfig(svmName string) string {
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

resource "netapp-ontap_name_services_dns_resource" "name_services_dns" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  svm_name = "%s"
  name_servers = ["1.1.1.1", "2.2.2.2"]
  dns_domains = ["foo.bar.com", "boo.bar.com"]
}
`, host, admin, password, svmName)
}
