package snapmirror_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSnapmirrorPolicyResourceAlias(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test snapmirror policy error
			{
				Config:      testAccSnapmirrorPolicyResourceBasicConfigAlias("non-existant"),
				ExpectError: regexp.MustCompile("2621462"),
			},
			// Test create snapmirror policy basic
			{
				Config: testAccSnapmirrorPolicyResourceBasicConfigAlias("ansibleSVM"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "name", "carchitestme4"),
				),
			},
			//  Test adding transfer_schedule
			{
				Config: testAccSnapmirrorPolicyResourceAddTransferScheduleBasicConfigAlias("ansibleSVM", "weekly"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "name", "carchitestme4"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "transfer_schedule_name", "weekly"),
				),
			},
			//  Test update transfer_schedule
			{
				Config: testAccSnapmirrorPolicyResourceAddTransferScheduleBasicConfigAlias("ansibleSVM", "daily"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "name", "carchitestme4"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "transfer_schedule_name", "daily"),
				),
			},
			// Test remove snapmirror policy transfer schedule
			{
				Config: testAccSnapmirrorPolicyResourceBasicConfigAlias("ansibleSVM"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "name", "carchitestme4"),
				),
			},
			// Test add snapmirror policy with comment and identity_preservation
			{
				Config: testAccSnapmirrorPolicyResourceConfigAlias("ansibleSVM", "test comment", "full"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "name", "carchitestme4"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "comment", "test comment"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "identity_preservation", "full"),
				),
			},
			// Test update snapmirror policy with comment and identity_preservation change
			{
				Config: testAccSnapmirrorPolicyResourceConfigAlias("ansibleSVM", "update comment", "exclude_network_config"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "name", "carchitestme4"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "comment", "update comment"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "identity_preservation", "exclude_network_config"),
				),
			},
			// Test update snapmirror policy with adding two retention rules
			{
				Config: testAccSnapmirrorPolicyResourceAddTwoRetentionConfigAlias("ansibleSVM", "update comment", "exclude_network_config", "weekly", 5),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "name", "carchitestme4"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "comment", "update comment"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "identity_preservation", "exclude_network_config"),
					// check number of reteion
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "retention.#", "2"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "retention.0.label", "hourly"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "retention.0.count", "7"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "retention.1.label", "weekly"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "retention.1.count", "5"),
				),
			},
			// Test update snapmirror policy with removing one retention rule
			{
				Config: testAccSnapmirrorPolicyResourceRemoveOneRetentionConfigAlias("ansibleSVM", "update comment", "exclude_network_config"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "name", "carchitestme4"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "comment", "update comment"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "identity_preservation", "exclude_network_config"),
					// check number of reteion
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "retention.#", "1"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "retention.0.label", "hourly"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.example", "retention.0.count", "7"),
				),
			},
			// Test create sync type snapmirror policy
			{
				Config: testAccSnapmirrorPolicyResourceSyncBasicConfigAlias("ansibleSVM", "test sync"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.sync_example", "name", "test_sync"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.sync_example", "comment", "test sync"),
				),
			},
			// Test update sync type snapmirror policy with changing comment
			{
				Config: testAccSnapmirrorPolicyResourceSyncBasicConfigAlias("ansibleSVM", "test update sync comment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.sync_example", "name", "test_sync"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.sync_example", "comment", "test update sync comment"),
				),
			},
			// Test update sync type snapmirror policy with adding a retention
			{
				Config: testAccSnapmirrorPolicyResourceSyncAddRetentionConfigAlias("ansibleSVM", "test add retenion in sync type"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.sync_example", "name", "test_sync"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.sync_example", "comment", "test add retenion in sync type"),
					// check number of reteion
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.sync_example", "retention.#", "1"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.sync_example", "retention.0.label", "daily"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy_resource.sync_example", "retention.0.count", "1"),
				),
			},
			// Test update sync type snapmirror policy with adding extra retention - max is 1
			{
				Config:      testAccSnapmirrorPolicyResourceSyncAddExtraRetentionConfigAlias("ansibleSVM", "test add extra retenion in sync type"),
				ExpectError: regexp.MustCompile("error updating sync snapshot policies"),
			},
		},
	})
}

func testAccSnapmirrorPolicyResourceBasicConfigAlias(svm string) string {
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

resource "netapp-ontap_snapmirror_policy_resource" "example" {
  cx_profile_name = "cluster4"
  name = "carchitestme4"
  svm_name = "%s"
  type = "async"
}`, host, admin, password, svm)
}

func testAccSnapmirrorPolicyResourceAddTransferScheduleBasicConfigAlias(svm string, transferScheduleName string) string {
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

resource "netapp-ontap_snapmirror_policy_resource" "example" {
  cx_profile_name = "cluster4"
  name = "carchitestme4"
  svm_name = "%s"
  type = "async"
  transfer_schedule_name = "%s"
}`, host, admin, password, svm, transferScheduleName)
}

func testAccSnapmirrorPolicyResourceConfigAlias(svm string, comment string, identityPreservation string) string {
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

resource "netapp-ontap_snapmirror_policy_resource" "example" {
  cx_profile_name = "cluster4"
  name = "carchitestme4"
  svm_name = "%s"
  comment = "%s"
  identity_preservation = "%s"
  type = "async"
}`, host, admin, password, svm, comment, identityPreservation)
}

func testAccSnapmirrorPolicyResourceAddTwoRetentionConfigAlias(svm string, comment string, identityPreservation string, label string, count int) string {
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

resource "netapp-ontap_snapmirror_policy_resource" "example" {
  cx_profile_name = "cluster4"
  name = "carchitestme4"
  svm_name = "%s"
  comment = "%s"
  identity_preservation = "%s"
  type = "async"
  retention = [
	{
		label = "hourly"
		count = 7
		creation_schedule_name = "hourly"
	},
	{
		label = "%s"
		count = %d
	},
  ]
}`, host, admin, password, svm, comment, identityPreservation, label, count)
}

func testAccSnapmirrorPolicyResourceRemoveOneRetentionConfigAlias(svm string, comment string, identityPreservation string) string {
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

resource "netapp-ontap_snapmirror_policy_resource" "example" {
  cx_profile_name = "cluster4"
  name = "carchitestme4"
  svm_name = "%s"
  comment = "%s"
  identity_preservation = "%s"
  type = "async"
  retention = [
	{
		label = "hourly"
		count = 7
		creation_schedule_name = "hourly"
	}
  ]
}`, host, admin, password, svm, comment, identityPreservation)
}

func testAccSnapmirrorPolicyResourceSyncBasicConfigAlias(svm string, comment string) string {
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

resource "netapp-ontap_snapmirror_policy_resource" "sync_example" {
  cx_profile_name = "cluster4"
  name = "test_sync"
  svm_name = "%s"
  type = "sync"
  sync_type = "sync"
  comment = "%s"
}`, host, admin, password, svm, comment)
}

func testAccSnapmirrorPolicyResourceSyncAddRetentionConfigAlias(svm string, comment string) string {
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

resource "netapp-ontap_snapmirror_policy_resource" "sync_example" {
  cx_profile_name = "cluster4"
  name = "test_sync"
  svm_name = "%s"
  type = "sync"
  sync_type = "sync"
  comment = "%s"
  retention = [
	{
		label = "daily"
		count = 1
	}
  ]
}`, host, admin, password, svm, comment)
}

func testAccSnapmirrorPolicyResourceSyncAddExtraRetentionConfigAlias(svm string, comment string) string {
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

resource "netapp-ontap_snapmirror_policy_resource" "sync_example" {
  cx_profile_name = "cluster4"
  name = "test_sync"
  svm_name = "%s"
  type = "sync"
  sync_type = "sync"
  comment = "%s"
  retention = [
	{
		label = "daily"
		count = 1
	},
	{
		label = "daily"
		count = 1
	}
  ]
}`, host, admin, password, svm, comment)
}
