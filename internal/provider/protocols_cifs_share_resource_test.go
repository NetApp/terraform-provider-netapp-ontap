package provider

// import (
// 	"fmt"
// 	"os"
// 	"regexp"
// 	"testing"

// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
// )

// func TestAccProtocolsCIFSShareResource(t *testing.T) {
// 	resource.Test(t, resource.TestCase{
// 		PreCheck:                 func() { testAccPreCheck(t) },
// 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// 		Steps: []resource.TestStep{
// 			// Test non existant SVM
// 			{
// 				Config:      testAccProtocolsCIFSShareResourceConfig("non-existant", "terraformTest4"),
// 				ExpectError: regexp.MustCompile("2621462"),
// 			},
// 			// Read testing
// 			{
// 				Config: testAccProtocolsCIFSShareResourceConfig("tfsvm", "acc_test_cifs_share"),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_share_resource.example", "name", "acc_test_cifs_share"),
// 					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_share_resource.example", "comment", "this is a comment"),
// 					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_share_resource.example", "continuously_available", "false"),
// 				),
// 			},
// 			{
// 				Config: testAccProtocolsCIFSShareResourceConfigUpdate("tfsvm", "acc_test_cifs_share"),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_share_resource.example", "name", "acc_test_cifs_share"),
// 					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_share_resource.example", "comment", "update comment"),
// 					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_share_resource.example", "continuously_available", "true"),
// 				),
// 			},
// 			// Test importing a resource
// 			{
// 				ResourceName:  "netapp-ontap_protocols_cifs_share_resource.example",
// 				ImportState:   true,
// 				ImportStateId: fmt.Sprintf("%s,%s,%s", "acc_test_cifs_share_import", "tfsvm", "clustercifs"),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr("netapp-ontap_protocols_cifs_share_resource.example", "name", "acc_test_cifs_share_import"),
// 				),
// 			},
// 		},
// 	})
// }

// func testAccProtocolsCIFSShareResourceConfig(svm, shareName string) string {

// 	if host == "" || admin == "" || password == "" {
// 		host = os.Getenv("TF_ACC_NETAPP_HOST_CIFS")
// 		admin = os.Getenv("TF_ACC_NETAPP_USER")
// 		password = os.Getenv("TF_ACC_NETAPP_PASS2")
// 	}
// 	if host == "" || admin == "" || password == "" {
// 		fmt.Println("TF_ACC_NETAPP_HOST_CIFS, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS2 must be set for acceptance tests")
// 		os.Exit(1)
// 	}
// 	return fmt.Sprintf(`
// provider "netapp-ontap" {
//  connection_profiles = [
//     {
//       name = "clustercifs"
//       hostname = "%s"
//       username = "%s"
//       password = "%s"
//       validate_certs = false
//     },
//   ]
// }

// resource "netapp-ontap_protocols_cifs_share_resource" "example" {
// 	cx_profile_name = "clustercifs"
//   	name = "%s"
//   	svm_name = "%s"
// 	path = "/acc_test_cifs_share_volume"
// 	acls = [
// 		{
// 	  		"permission": "read",
// 	  		"type": "windows",
// 	  		"user_or_group": "Everyone"
// 		}
// 	]
//  	comment = "this is a comment"
// }`, host, admin, password, shareName, svm)
// }

// func testAccProtocolsCIFSShareResourceConfigUpdate(svm, volName string) string {
// 	if host == "" || admin == "" || password == "" {
// 		host = os.Getenv("TF_ACC_NETAPP_HOST2")
// 		admin = os.Getenv("TF_ACC_NETAPP_USER")
// 		password = os.Getenv("TF_ACC_NETAPP_PASS2")
// 	}
// 	if host == "" || admin == "" || password == "" {
// 		fmt.Println("TF_ACC_NETAPP_HOST_CIFS, TF_ACC_NETAPP_USER, and TF_ACC_NETAPP_PASS2 must be set for acceptance tests")
// 		os.Exit(1)
// 	}
// 	return fmt.Sprintf(`
// provider "netapp-ontap" {
//  connection_profiles = [
//     {
//       name = "clustercifs"
//       hostname = "%s"
//       username = "%s"
//       password = "%s"
//       validate_certs = false
//     },
//   ]
// }

// resource "netapp-ontap_protocols_cifs_share_resource" "example" {
//   cx_profile_name = "clustercifs"
//   name = "%s"
//   svm_name = "%s"
//   path = "/acc_test_cifs_share_volume"
//   acls = [
// 	  {
// 			"permission": "read",
// 			"type": "windows",
// 			"user_or_group": "Everyone"
// 	  }
//   ]
//   comment = "update comment"
//   continuously_available = true
// }`, host, admin, password, volName, svm)
// }
