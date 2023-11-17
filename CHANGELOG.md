
## 1.1.0 ()

ENHANCEMENTS:
* **netapp-ontap_cluster_licensing_license_resource**: Add support for import ([#30](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/30))
* **netapp-ontap_storage_aggregate_resource**: Add support for import ([#39](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/39))
* **netapp-ontap_storage_volume_resource**: Add support for import ([#41](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/41))
* **netapp-ontap_protocols_nfs_service_resource**: Add support for import ([#36](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/36))


## 1.0.2 (2023-11-17)
* 1.0.1 did not deploy correctly 1.0.2 fixes that. 


## 1.0.1 (2023-11-17)

BUG FIXES:
* netapp-ontap_name_services_dns_resource: Fixed and Documented Import ([#63](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/63))
* netapp-ontap_cluster_data_source, netapp-ontap_snapmirrors_data_source, netapp-ontap_networking_ip_route_resource and netapp-ontap_sotrage_volume_resource: Fix documentation ([#70](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/70))


## 1.0.0 (2023-11-06)

FEATURES:
* **New Data Source:** `netapp-ontap_cluster_data_source`
* **New Data Source:** `netapp-ontap_cluster_licensing_license_data_source`
* **New Data Source:** `netapp-ontap_cluster_licensing_licenses_data_source`
* **New Data Source:** `netapp-ontap_cluster_schedule_data_source`
* **New Data Source:** `netapp-ontap_cluster_schedules_data_source`
* **New Data Source:** `netapp-ontap_ip_interface_data_source`
* **New Data Source:** `netapp-ontap_ip_interfaces_data_source`
* **New Data Source:** `netapp-ontap_name_services_dns_data_source`
* **New Data Source:** `netapp-ontap_name_services_dnss_data_source`
* **New Data Source:** `netapp-ontap_networking_ip_route_data_source`
* **New Data Source:** `netapp-ontap_networking_ip_routes_data_source`
* **New Data Source:** `netapp-ontap_protcols_nfs_service_data_source`
* **New Data Source:** `netapp-ontap_protcols_nfs_services_data_source`
* **New Data Source:** `netapp-ontap_protocols_nfs_export_policies_data_source`
* **New Data Source:** `netapp-ontap_protocols_nfs_export_policy_data_source`
* **New Data Source:** `netapp-ontap_protocols_nfs_export_policy_rule_data_source`
* **New Data Source:** `netapp-ontap_snapmirror_policies_data_source`
* **New Data Source:** `netapp-ontap_snapmirror_policy_data_source`
* **New Data Source:** `netapp-ontap_storage_aggregate_data_source`
* **New Data Source:** `netapp-ontap_storage_aggregates_data_source`
* **New Data Source:** `netapp-ontap_storage_snapshot_policies_data_source`
* **New Data Source:** `netapp-ontap_storage_snapshot_policy_data_source`
* **New Data Source:** `netapp-ontap_storage_volume_data_source`
* **New Data Source:** `netapp-ontap_storage_volumes_data_source`
* **New Data Source:** `netapp-ontap_storage_volume_snapshot_data_source`
* **New Data Source:** `netapp-ontap_svm_data_source`
* **New Data Source:** `netapp-ontap_svms_data_source`
* **New Resource:** `netapp-ontap_cluster_licensing_license_resource`
* **New Resource:** `netapp-ontap_cluster_schedule_resource`
* **New Resource:** `netapp-ontap_networking_ip_interface_resource`
* **New Resource:** `netapp-ontap_name_services_dns_resource`
* **New Resource:** `netapp-ontap_networking_ip_route_resource`
* **New Resource:** `netapp-ontap_protocols_nfs_export_policy_resource`
* **New Resource:** `netapp-ontap_protocols_nfs_export_policy_rule_resource`
* **New Resource:** `netapp-ontap_snapmirror_resource`
* **New Resource:** `netapp-ontap_snapmirror_policy_resource`
* **New Resource:** `netapp-ontap_storage_aggregate_resource`
* **New Resource:** `netapp-ontap_storage_snapshot_policy_resource`
* **New Resource:** `netapp-ontap_storage_volume_resource`
* **New Resource:** `netapp-ontap_storage_volume_snapshot_resource`
* **New Resource:** `netapp-ontap_svm_resource`
