# 1.2.0 ()

BREAKING CHANGES:
* **Rename Resource:** `netapp-ontap_cluster_peers` is now renamed to `netapp-ontap_cluster_peer`
* **Rename Resource:** `netapp-ontap_cifs_local_group_member` is now renamed to `netapp-ontap_cifs_local_group_members`
* **Rename Resource:** `netapp-ontap_cifs_user_group_privilege` is now renamed to `netapp-ontap_cifs_user_group_privileges`
* **Rename Resource:** `netapp-ontap_san_lun-maps` is now renamed to `netapp-ontap_san_lun-map`
* **Rename Resource:** `netapp-ontap_security_accounts` is now renamed to `netapp-ontap_security_account`
* **Rename Resource:** `netapp-ontap_security_login_messages` is now renamed to `netapp-ontap_security_login_message`
* **Rename Resource:** `netapp-ontap_security_roles` is now renamed to `netapp-ontap_security_role`


FEATURES:
* **provider**: add `aws_lambda` option. ([#262](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/262))
* **New Data Source:** `netapp-ontap_volumes_files` ([#8](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/8))
* **New Data Source:** `netapp-ontap_qos_policy` ([#77](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/77))
* **New Data Source:** `netapp-ontap_qos_policies` ([#77](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/77))
* **New Data Source:** `netapp-ontap_quota_rules` ([#135](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/135))
* **New Data Source:** `netapp-ontap_quota_rule` ([#135](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/135))
* **New Data Source:** `netapp-ontap_security_role` ([#139](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/139))
* **New Data Source:** `netapp-ontap_security_roles` ([#139](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/139))
* **New Data Source:** `netapp-ontap_security_login_message` ([#17](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/17))
* **New Data Source:** `netapp-ontap_security_login_messages` ([#17](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/17))
* **New Data Source:** `netapp-ontap_storage_qtree` ([#83](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/83))
* **New Data Source:** `netapp-ontap_storage_qtrees` ([#83](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/83))
* **New Data Source:** `netapp-ontap_volume_efficiency_policy` ([#81](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/81))
* **New Data Source:** `netapp-ontap_volume_efficiency_policies` ([#81](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/81))
* **New Resource:** `netapp-ontap_volume_efficiency_policies` ([#80](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/80))
* **New Resource:** `netapp-ontap_quota_rules` ([#136](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/136))
* **New Resource:** `netapp-ontap_volumes_files` ([#5](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/5))
* **New Resource:** `netapp-ontap_security_roles` ([#140](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/140))
* **New Resource:** `netapp-ontap_storage_qtrees` ([#82](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/82))
* **New Resource:** `netapp-ontap_qos_policies` ([#76](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/76))
* **New Resource:** `netapp-security_login_messages` ([#18](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/18))

ENHANCEMENTS:
* **netapp-ontap_lun**: added `size_unit` option. ([#227](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/227))
* **netapp-ontap_security_account**: Add support for import and update ([#243](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/243))

## 1.1.4 (2024-09-05)

DOC FIXES:
* **netapp-ontap_storage_flexcache_resource**: Fixed Page display issue ([[#271](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/271)])
* **netapp-ontap_networking_ip_interface_resource**: Include min version for metrics ([[#265](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/265)])

BUG FIXES:
* **netapp-ontap_cluster_data_source: fix on nodes to show multiple elements ([#264](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/264))

## 1.1.3 (2024-08-08)

BUG FIXES:
* **netapp-ontap_protocols_cifs_service_resource**: fixed on attribute checking ([#250](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/250))
* **netapp-ontap_protocols_cifs_share_resource** :`acls` unable to update acls ([#236](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/236))
* **netapp-ontap_protocols_san_igroups_resource**: fixed bug nil pointer dereference ([#247](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/247))

## 1.1.2 (2024-06-03)

ENHANCEMENTS:
* **netapp-ontap_storage_lun_resource**, **netapp-ontap_storage_lun_data_source**, **netapp-ontap_storage_luns_data_source**: added `serial_number` option to get info in state file ([#207](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/207))

BUG FIXES:
* **netapp-ontap_storage_lun_resource**, **netapp-ontap_storage_lun_data_source**, **netapp-ontap_storage_luns_data_source**: fixed `name` and `logical_unit` options as separate inputs for resource and info in state file gets added separately as well ([#215](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/215))
* added guide for changing NetApp Ontap Provider from one minor version to another.
* corrected typos in the CHANGELOG.

BUG FIXES:
* netapp-ontap_storage_flexcache_resource: Fixed junction_path bug ([#223](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/223))

## 1.1.1 (2024-05-15)

* added missing resources in changelog.
* corrected typos in the documentation.

## 1.1.0 (2024-05-08)

FEATURES:
* **New Data Source:** `netapp-ontap_protocols_cifs_local_group_data_source` ([#54](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/54))
* **New Data Source:** `netapp-ontap_protocols_cifs_local_groups_data_source` ([#54](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/54))
* **New Data Source:** `netapp-ontap_cluster_peer_data_source` ([#50](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/50))
* **New Data Source:** `netapp-ontap_cluster_peers_data_source` ([#50](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/50))
* **New Data Source:** `netapp-ontap_protocols_cifs_local_user_data_source` ([#55](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/55))
* **New Data Source:** `netapp-ontap_protocols_cifs_local_users_data_source` ([#55](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/55))
* **New Data Source:** `netapp-ontap_security_account_data_source` ([#22](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/22))
* **New Data Source:** `netapp-ontap_security_accounts_data_source` ([#22](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/22))
* **New Data Source:** `netapp-ontap_protocols_cifs_user_group_privilege_data_source` ([#57](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/57))
* **New Data Source:** `netapp-ontap_protocols_cifs_user_group_privileges_data_source` ([#57](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/57))
* **New Data Source:** `netapp-ontap_storage_lun_data_source` ([#12](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/12))
* **New Data Source:** `netapp-ontap_storage_luns_data_source` ([#12](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/12))
* **New Data Source:** `netapp-ontap_protocols_cifs_local_group_member_data_source` ([#122](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/122))
* **New Data Source:** `netapp-ontap_protocols_cifs_local_group_members_data_source` ([#122](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/122))
* **New Data Source:** `netapp-ontap_svm_peer_data_source` ([#52](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/52))
* **New Data Source:** `netapp-ontap_svm_peers_data_source` ([#52](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/52))
* **New Data Source:** `netapp-ontap_protocols_cifs_server_data_source` ([#24](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/24))
* **New Data Source:** `netapp-ontap_protocols_cifs_servers_data_source` ([#24](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/24))
* **New Data Source:** `netapp-ontap_storage_flexcache_data_source` ([#47](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/47))
* **New Data Source:** `netapp-ontap_storage_flexcaches_data_source` ([#47](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/47))
* **New Data Source:** `netapp-ontap_name_services_ldap_data_source` ([#26](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/26))
* **New Data Source:** `netapp-ontap_name_services_ldaps_data_source` ([#26](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/26))
* **New Data Source:** `netapp-ontap_protocols_cifs_share_data_source` ([#28](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/28))
* **New Data Source:** `netapp-ontap_protocols_cifs_shares_data_source` ([#28](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/28))
* **New Data Source:** `netapp-ontap_protocols_san_lun-map_data_source` ([#14](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/14))
* **New Data Source:** `netapp-ontap_protocols_san_lun-maps_data_source` ([#14](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/14))
* **New Resource:** `netapp-ontap_protocols_cifs_local_group_resource` ([#53](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/53))
* **New Resource:** `netapp-ontap_protocols_cifs_local_user_resource` ([#56](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/56))
* **New Resource:** `netapp-ontap_protocols_cifs_user_group_privilege_resource` ([#58](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/58))
* **New Resource:** `netapp-ontap_svm_peers_resource` ([#51](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/51))
* **New Resource:** `netapp-ontap_protocols_cifs_user_group_member_resource` ([#123](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/123))
* **New Resource:** `netapp-ontap_storage_flexcache_resource` ([#46](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/46))
* **New Resource:** `netapp-ontap_protocols_san_lun-maps_resource` ([#13](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/13))
* **New Resource:** `netapp-ontap_name_services_ldap_resource` ([#25](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/25))
* **New Resource:** `netapp-ontap_protocols_cifs_service_resource` ([#23](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/23))
* **New Resource:** `netapp-ontap_protocols_cifs_share_resource` ([#27](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/27))
* **New Resource:** `netapp-ontap_protocols_san_igroup_resource` ([#9](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/9))
* **New Resource:** `netapp-ontap_cluster_resource` ([#15](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/15))
* **New Resource:** `netapp-ontap_cluster_peers_resource` ([#49](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/49))
* **New Resource:** `netapp-ontap_storage_lun_resource` ([#11](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/11))
* **New Resource:** `netapp-ontap_security_account_resource` ([#21](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/21))


ENHANCEMENTS:
* **netapp-ontap_protocols_nfs_export_policy_resource**: Add support for import ([#34](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/34))
* **netapp-ontap_cluster_licensing_license_resource**: Add support for import ([#30](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/30))
* **netapp-ontap_storage_aggregate_resource**: Add support for import ([#39](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/39))
* **netapp-ontap_storage_volume_resource**: Add support for import ([#41](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/41))
* **netapp-ontap_protocols_nfs_service_resource**: Add support for import ([#36](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/36))
* **netapp-ontap_svm_resource**: Add support for import ([#6](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/6))
* **netapp-ontap_storage_volume_snapshot_resource**: Add support for import ([#42](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/42))
* **netapp-ontap_cluster_schedule_resource**: Add support for import ([#31](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/31))
* **netapp-ontap_storage_snapshot_policy_resource**: Add support for import ([#40](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/40))
* **netapp-ontap_snapmirror_resource**: Add support for modify ([#45](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/45))
* **netapp-ontap_snapmirror_resource**: Add support for import ([#37](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/37))
* **netapp-ontap_snapmirror_policy**: Add support for import ([#38](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/38))
* **netapp-ontap_networking_ip_interface_resource**: Add support for import ([#32](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/32))
* **netapp-ontap_protocols_nfs_export_policy_rule_resource**: Add support for import ([#35](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/35))
* **netapp-ontap_networking_ip_route_resource**: Add support for import ([#33](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/33))
* **netapp-ontap_cluster** Add contact, locaton, dns_domains, name_servers, timezone, certificate, ntp_servers, management_interfaces options ([#16](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/16))

BUG FIXES:
* **netapp-ontap_protocols_nfs_service**: Fixed issue with version check failing for minor versions
* netapp-ontap resource and data source documentation update ([#169](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/169))


## 1.0.3 (2023-12-05)
* netapp-ontap_name_services_dns_resource: Fixed missing ID on create ([#99](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/63))

## 1.0.2 (2023-11-17)
* 1.0.1 did not deploy correctly 1.0.2 fixes that. 


## 1.0.1 (2023-11-17)

BUG FIXES:
* netapp-ontap_name_services_dns_resource: Fixed and Documented Import ([#63](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/63))
* netapp-ontap_cluster_data_source, netapp-ontap_snapmirrors_data_source, netapp-ontap_networking_ip_route_resource and netapp-ontap_sotrage_volume_resource: Fix documentation ([#70](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/70))
* netapp-ontap_name_services_ldap_resource: Fixed and Document with the version check and workflow ([#163](https://github.com/NetApp/terraform-provider-netapp-ontap/issues/163))


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
