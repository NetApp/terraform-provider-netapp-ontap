package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNFSExportPolicyRuleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccNFSExportPolicyRuleResourceConfigMissingVars("non-existant"),
				ExpectError: regexp.MustCompile("Missing required argument"),
			},
			// create with basic argument
			{
				Config: testAccNFSExportPolicyRuleResourceConfig("carchi-test", "default"),
				Check: resource.ComposeTestCheckFunc(
					// check default values
					resource.TestCheckResourceAttr("netapp-ontap_protocols_nfs_export_policy_rule_resource.example1", "allow_suid", "true"),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_protocols_nfs_export_policy_rule_resource.example1", "protocols.*", "any"),
					// check id
					resource.TestMatchResourceAttr("netapp-ontap_protocols_nfs_export_policy_rule_resource.example1", "id", regexp.MustCompile(`carchi-test_default_`)),
				),
			},
			// update test
			{
				Config: testAccNFSExportPolicyRuleResourceConfigUpdateProtocolsROrule("carchi-test", "default"),
				Check: resource.ComposeTestCheckFunc(
					// check default values
					resource.TestCheckResourceAttr("netapp-ontap_protocols_nfs_export_policy_rule_resource.example1", "allow_suid", "true"),
					// check if the modification successful
					resource.TestCheckTypeSetElemAttr("netapp-ontap_protocols_nfs_export_policy_rule_resource.example1", "protocols.*", "nfs3"),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_protocols_nfs_export_policy_rule_resource.example1", "ro_rule.*", "krb5i"),
					// check id
					resource.TestMatchResourceAttr("netapp-ontap_protocols_nfs_export_policy_rule_resource.example1", "id", regexp.MustCompile(`carchi-test_default_`)),
				),
			},
			// Test importing a resource
			{
				ResourceName:  "netapp-ontap_protocols_nfs_export_policy_rule_resource.example1",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s,%s,%s,%s", "1", "carchi-test", "default", "cluster4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_protocols_nfs_export_policy_rule_resource.example1", "svm_name", "carchi-test"),
					resource.TestCheckResourceAttr("netapp-ontap_protocols_nfs_export_policy_rule_resource.example1", "export_policy_name", "default"),
					resource.TestCheckResourceAttr("netapp-ontap_protocols_nfs_export_policy_rule_resource.example1", "allow_suid", "true"),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_protocols_nfs_export_policy_rule_resource.example1", "protocols.*", "nfs3"),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_protocols_nfs_export_policy_rule_resource.example1", "ro_rule.*", "krb5i"),
					resource.TestCheckTypeSetElemAttr("netapp-ontap_protocols_nfs_export_policy_rule_resource.example1", "rw_rule.*", "any"),
					// check id
					resource.TestMatchResourceAttr("netapp-ontap_protocols_nfs_export_policy_rule_resource.example1", "id", regexp.MustCompile(`carchi-test_default_`)),
				),
			},
		},
	})
}

func testAccNFSExportPolicyRuleResourceConfigMissingVars(svmName string) string {
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

resource "netapp-ontap_protocols_nfs_export_policy_rule_resource" "example" {
  cx_profile_name = "cluster4"
  svm_name = "%s"
}
`, host, admin, password, svmName)
}

func testAccNFSExportPolicyRuleResourceConfig(svmName string, exportPolicyName string) string {
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

resource "netapp-ontap_protocols_nfs_export_policy_rule_resource" "example1" {
  cx_profile_name = "cluster4"
  svm_name = "%s"
  export_policy_name = "%s"
  clients_match = ["0.0.0.0/0"]
  ro_rule = ["any"]
  rw_rule = ["any"]
}
`, host, admin, password, svmName, exportPolicyName)
}

// update protocols and ro_rule
func testAccNFSExportPolicyRuleResourceConfigUpdateProtocolsROrule(svmName string, exportPolicyName string) string {
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

resource "netapp-ontap_protocols_nfs_export_policy_rule_resource" "example1" {
  cx_profile_name = "cluster4"
  svm_name = "%s"
  export_policy_name = "%s"
  protocols = ["nfs3","nfs"]
  clients_match = ["0.0.0.0/0"]
  ro_rule = ["krb5","krb5i"]
  rw_rule = ["any"]
}
`, host, admin, password, svmName, exportPolicyName)
}
