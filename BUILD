load("@bazel_gazelle//:def.bzl", "gazelle")
load("@rules_oci//oci:defs.bzl", "oci_push")
load(
    "//:config.bzl",
    "deploy_env",
    "gcp_project_id",
    "gcp_project_number",
    "gcp_region",
    "gcp_zone",
    "registry_api_image",
    "registry_jupyter_image",
    "registry_monitor_image",
    "registry_user_base_image",
)

# Generate app config file
# TODO: this is temporary solution. Please remove in the future.
sh_binary(
    name = "generate_app_config_sh",
    srcs = [":generate_app_config_file.sh"],
    args = [
        gcp_project_id,
        gcp_project_number,
        deploy_env,
        gcp_region,
        gcp_zone,
    ],
)

genrule(
    name = "generate_app_config",
    srcs = [],
    outs = ["config.yaml"],
    cmd = "$(location :generate_app_config_sh) {} {} {} {} {} > $(OUTS)".format(
        gcp_project_id,
        gcp_project_number,
        deploy_env,
        gcp_region,
        gcp_zone,
    ),
    tools = [":generate_app_config_sh"],
    visibility = ["//visibility:public"],
)

# gazelle:prefix github.com/manatee-project/manatee
gazelle(name = "gazelle")

# push images
oci_push(
    name = "push_dcr_api_image",
    image = "//app/dcr_api:image",
    remote_tags = ["latest"],
    repository = registry_api_image,
)

oci_push(
    name = "push_dcr_monitor_image",
    image = "//app/dcr_monitor:image",
    remote_tags = ["latest"],
    repository = registry_monitor_image,
)

oci_push(
    name = "push_jupyterlab_image",
    image = "//app/jupyterlab_manatee:image",
    remote_tags = ["latest"],
    repository = registry_jupyter_image,
)

# push user images
oci_push(
    name = "push_dcr_tee_image",
    image = "//app/dcr_tee:image",
    remote_tags = ["latest"],
    repository = registry_user_base_image,
)
