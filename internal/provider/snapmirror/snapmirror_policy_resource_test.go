package snapmirror_test

import (
	"fmt"
	ntest "github.com/netapp/terraform-provider-netapp-ontap/internal/provider"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSnapmirrorPolicyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { ntest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: ntest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test snapmirror policy error
			{
				Config:      testAccSnapmirrorPolicyResourceBasicConfig("non-existant"),
				ExpectError: regexp.MustCompile("2621462"),
			},
			// Test create snapmirror policy basic
			{
				Config: testAccSnapmirrorPolicyResourceBasicConfig("ansibleSVM"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "name", "carchitestme4"),
				),
			},
			//  Test adding transfer_schedule
			{
				Config: testAccSnapmirrorPolicyResourceAddTransferScheduleBasicConfig("ansibleSVM", "weekly"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "name", "carchitestme4"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "transfer_schedule_name", "weekly"),
				),
			},
			//  Test update transfer_schedule
			{
				Config: testAccSnapmirrorPolicyResourceAddTransferScheduleBasicConfig("ansibleSVM", "daily"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "name", "carchitestme4"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "transfer_schedule_name", "daily"),
				),
			},
			// Test remove snapmirror policy transfer schedule
			{
				Config: testAccSnapmirrorPolicyResourceBasicConfig("ansibleSVM"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "name", "carchitestme4"),
				),
			},
			// Test add snapmirror policy with comment and identity_preservation
			{
				Config: testAccSnapmirrorPolicyResourceConfig("ansibleSVM", "test comment", "full"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "name", "carchitestme4"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "comment", "test comment"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "identity_preservation", "full"),
				),
			},
			// Test update snapmirror policy with comment and identity_preservation change
			{
				Config: testAccSnapmirrorPolicyResourceConfig("ansibleSVM", "update comment", "exclude_network_config"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "name", "carchitestme4"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "comment", "update comment"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "identity_preservation", "exclude_network_config"),
				),
			},
			// Test update snapmirror policy with adding two retention rules
			{
				Config: testAccSnapmirrorPolicyResourceAddTwoRetentionConfig("ansibleSVM", "update comment", "exclude_network_config", "weekly", 5),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "name", "carchitestme4"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "comment", "update comment"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "identity_preservation", "exclude_network_config"),
					// check number of reteion
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "retention.#", "2"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "retention.0.label", "hourly"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "retention.0.count", "7"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "retention.1.label", "weekly"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "retention.1.count", "5"),
				),
			},
			// Test update snapmirror policy with removing one retention rule
			{
				Config: testAccSnapmirrorPolicyResourceRemoveOneRetentionConfig("ansibleSVM", "update comment", "exclude_network_config"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "name", "carchitestme4"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "comment", "update comment"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "identity_preservation", "exclude_network_config"),
					// check number of reteion
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "retention.#", "1"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "retention.0.label", "hourly"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.example", "retention.0.count", "7"),
				),
			},
			// Test create sync type snapmirror policy
			{
				Config: testAccSnapmirrorPolicyResourceSyncBasicConfig("ansibleSVM", "test sync"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.sync_example", "name", "test_sync"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.sync_example", "comment", "test sync"),
				),
			},
			// Test update sync type snapmirror policy with changing comment
			{
				Config: testAccSnapmirrorPolicyResourceSyncBasicConfig("ansibleSVM", "test update sync comment"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.sync_example", "name", "test_sync"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.sync_example", "comment", "test update sync comment"),
				),
			},
			// Test update sync type snapmirror policy with adding a retention
			{
				Config: testAccSnapmirrorPolicyResourceSyncAddRetentionConfig("ansibleSVM", "test add retenion in sync type"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.sync_example", "name", "test_sync"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.sync_example", "comment", "test add retenion in sync type"),
					// check number of reteion
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.sync_example", "retention.#", "1"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.sync_example", "retention.0.label", "daily"),
					resource.TestCheckResourceAttr("netapp-ontap_snapmirror_policy.sync_example", "retention.0.count", "1"),
				),
			},
			// Test update sync type snapmirror policy with adding extra retention - max is 1
			{
				Config:      testAccSnapmirrorPolicyResourceSyncAddExtraRetentionConfig("ansibleSVM", "test add extra retenion in sync type"),
				ExpectError: regexp.MustCompile("error updating sync snapshot policies"),
			},
		},
	})
}

func testAccSnapmirrorPolicyResourceBasicConfig(svm string) string {
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

resource "netapp-ontap_snapmirror_policy" "example" {
  cx_profile_name = "cluster4"
  name = "carchitestme4"
  svm_name = "%s"
  type = "async"
}`, host, admin, password, svm)
}

func testAccSnapmirrorPolicyResourceAddTransferScheduleBasicConfig(svm string, transferScheduleName string) string {
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

resource "netapp-ontap_snapmirror_policy" "example" {
  cx_profile_name = "cluster4"
  name = "carchitestme4"
  svm_name = "%s"
  type = "async"
  transfer_schedule_name = "%s"
}`, host, admin, password, svm, transferScheduleName)
}

func testAccSnapmirrorPolicyResourceConfig(svm string, comment string, identityPreservation string) string {
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

resource "netapp-ontap_snapmirror_policy" "example" {
  cx_profile_name = "cluster4"
  name = "carchitestme4"
  svm_name = "%s"
  comment = "%s"
  identity_preservation = "%s"
  type = "async"
}`, host, admin, password, svm, comment, identityPreservation)
}

func testAccSnapmirrorPolicyResourceAddTwoRetentionConfig(svm string, comment string, identityPreservation string, label string, count int) string {
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

resource "netapp-ontap_snapmirror_policy" "example" {
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

func testAccSnapmirrorPolicyResourceRemoveOneRetentionConfig(svm string, comment string, identityPreservation string) string {
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

resource "netapp-ontap_snapmirror_policy" "example" {
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

func testAccSnapmirrorPolicyResourceSyncBasicConfig(svm string, comment string) string {
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

resource "netapp-ontap_snapmirror_policy" "sync_example" {
  cx_profile_name = "cluster4"
  name = "test_sync"
  svm_name = "%s"
  type = "sync"
  sync_type = "sync"
  comment = "%s"
}`, host, admin, password, svm, comment)
}

func testAccSnapmirrorPolicyResourceSyncAddRetentionConfig(svm string, comment string) string {
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

resource "netapp-ontap_snapmirror_policy" "sync_example" {
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

func testAccSnapmirrorPolicyResourceSyncAddExtraRetentionConfig(svm string, comment string) string {
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

resource "netapp-ontap_snapmirror_policy" "sync_example" {
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
