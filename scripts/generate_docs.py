import subprocess
import os
import sys

# TO find the correct Category, check REST API to see what main header this API lives under

CATEGORIES = {
    'application': [],
    'cloud': [],
    'cluster': [
        "cluster_resource.md",
        "cluster_data_source.md",
        "cluster_license_data_source.md",
        "cluster_licenses_data_source.md",
        "cluster_licensing_license_resource.md",
        "cluster_peer_data_source.md",
        "cluster_peers_data_source.md",
        "cluster_peer_resource.md",
        "cluster_schedule_data_source.md",
        "cluster_schedules_data_source.md",
        "cluster_schedule_resource.md"],
    'nas': [
        "protocols_cifs_local_group_data_source.md",
        "protocols_cifs_local_groups_data_source.md",
        "protocols_cifs_local_group_resource.md",
        "protocols_cifs_local_group_member_data_source.md",
        "protocols_cifs_local_group_members_data_source.md",
        "protocols_cifs_local_group_members_resource.md",
        "protocols_cifs_local_user_data_source.md",
        "protocols_cifs_local_users_data_source.md",
        "protocols_cifs_local_user_resource.md",
        "protocols_cifs_service_data_source.md",
        "protocols_cifs_services_data_source.md",
        "protocols_cifs_service_resource.md",
        "protocols_cifs_share_data_source.md",
        "protocols_cifs_shares_data_source.md",
        "protocols_cifs_share_resource.md",
        "protocols_cifs_user_group_privileges_data_source.md",
        "protocols_cifs_user_group_privilege_data_source.md",
        "protocols_cifs_user_group_privilege_resource.md",
        "protocols_nfs_export_policies_data_source.md",
        "protocols_nfs_export_policy_data_source.md",
        "protocols_nfs_export_policy_resource.md",
        "protocols_nfs_export_policy_rule_data_source.md",
        "protocols_nfs_export_policy_rules_data_source.md",
        "protocols_nfs_export_policy_rule_resource.md",
        "protocols_nfs_service_data_source.md",
        "protocols_nfs_services_data_source.md",
        "protocols_nfs_service_resource.md"],
    'name-services': [
        "name_services_dns_data_source.md",
        "name_services_dnss_data_source.md",
        "name_services_dns_resource.md"
        "name_services_ldap_data_source.md",
        "name_services_ldaps_data_source.md",
        "name_services_ldap_resource.md",
    ],
    'ndmp': [],
    'networking': [
        "network_ip_interface_data_source.md",
        "network_ip_interfaces_data_source.md",
        "network_ip_interface_resource.md",
        "network_ip_route_data_source.md",
        "network_ip_routes_data_source.md",
        "network_ip_route_resource.md"],
    'nvme': [],
    'object-store': [],
    'san': [
        "protocols_san_igroup_data_source.md",
        "protocols_san_igroups_data_source.md",
        "protocols_san_igroup_resource.md",
        "protocols_san_lun-map_data_source.md",
        "protocols_san_lun-maps_data_source.md",
        "protocols_san_lun-map_resource.md",
    ],
    'security': [
        "security_account_data_source.md",
        "security_accounts_data_source.md",
        "security_account_resource.md",
        "security_login_message_data_source.md",
        "security_login_message_resource.md",
        "security_login_messages_data_source.md",
        "security_role_data_source.md",
        "security_roles_data_source.md",
        "security_roles_resource.md",
        "security_login_message_resource.md",
        "security_certificate_data_source.md",
        "security_certificates_data_source.md",

    ],
    'snaplock': [],
    'snapmirror': [
        "snapmirror_policies_data_source.md",
        "snapmirror_policy_data_source.md",
        "snapmirror_policy_resource.md",
        "snapmirror_data_source.md",
        "snapmirrors_data_source.md",
        "snapmirror_resource.md"],
    'storage': [
        "storage_aggregate_data_source.md",
        "storage_aggregates_data_source.md",
        "storage_aggregate_resource.md",
        "storage_flexcache_data_source.md",
        "storage_flexcaches_data_source.md",
        "storage_flexcache_resource.md",
        "storage_lun_data_source.md",
        "storage_luns_data_source.md",
        "storage_lun_resource.md",
        "storage_qos_policies_data_source.md",
        "storage_qos_policy_data_source.md",
        "storage_qos_policy_resoruce.md",
        "storage_quota_rule_data_source.md",
        "storage_quota_rules_data_source.md",
        "storage_quota_rule_resource.md",
        "storage_qtree_data_source.md",
        "storage_qtrees_data_source.md",
        "storage_qtree_resource.md",
        "storage_snapshot_policies_data_source.md",
        "storage_snapshot_policy_data_source.md",
        "storage_snapshot_policy_resource.md",
        "storage_volume_data_source.md",
        "storage_volumes_data_source.md",
        "storage_volumes_files_data_source.md",
        "storage_volume_resource.md",
        "storage_volume_snapshot_data_source.md",
        "storage_volume_snapshots_data_source.md"
        "storage_volume_snapshot_resource.md"],
    'support': [],
    'svm': ["svm_resource.md",
            "svm_peer_resource.md",
            "svm_peer_data_source.md",
            "svm_peers_data_source.md",],
}


def main():
    print("===== Generating docs =====")
    generate_doc()
    remove_example()
    print("===== Adding Categories =====")
    add_catagories()
    print("===== Validate =====")
    validate()
    print("===== Errors =====")
    issue = warn_missing_category(["docs/data-sources/", "docs/resources/"])
    if issue:
        sys.exit(1)


def generate_doc():
    cmd_str = "go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate"
    subprocess.run(cmd_str, shell=True)


def validate():
    cmd_str = "go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs validate"
    subprocess.run(cmd_str, shell=True)


def remove_example():
    files = ['docs/data-sources/example.md', 'docs/resources/example.md']
    for file_path in files:
        if os.path.exists(file_path):
            os.remove(file_path)


def add_catagories():
    for category in CATEGORIES:
        for page in CATEGORIES[category]:
            if 'data_source' in page:
                update_datasource(page, category)
            if 'resource' in page:
                update_resouces(page, category)


def update_datasource(page, category):
    path = "docs/data-sources/" + page
    update_md_file(path, category)


def update_resouces(page, category):
    path = "docs/resources/" + page
    update_md_file(path, category)


def update_md_file(path, category):
    print("Updating %s" % path)
    try:
        with open(path) as f:
            lines = f.readlines()
        for i, line in enumerate(lines):
            if line.startswith('subcategory: "'):
                split_line = line.split('subcategory: "')
                new_line = split_line[0] + 'subcategory: "' + category + split_line[1]
                lines[i] = new_line
                break
        with open(path, 'w') as f:
            f.writelines(lines)
    except:
        return

def warn_missing_category(directory_paths):
    issue = False
    for directory_path in directory_paths:
        full_path = os.path.join(os.getcwd(), directory_path)
        for filename in os.listdir(full_path):
            if os.path.isfile(os.path.join(full_path, filename)):
                with open(os.path.join(full_path, filename), 'r') as f:
                    file_content = f.read()
                    if 'subcategory: ""' in file_content:
                        print('%s is missing a category' % filename)
                        issue = True
    return issue


if __name__ == "__main__":
    main()