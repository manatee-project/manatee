# NOTE: this is a hacky way to load the configs, because it's not actually a .bzl file
# In the future, we need to remove the separate env file and rely on a single config file.
load("//:env.bzl", "env", "project_id", "region", "zone")

# reassign external variables to bzl variables here.
gcp_project_id = project_id
deploy_env = env
gcp_region = region
gcp_zone = zone

# artifact registries
registry_api_image = "us-docker.pkg.dev/{}/dcr-{}-images/data-clean-room-api".format(gcp_project_id, deploy_env)
registry_monitor_image = "us-docker.pkg.dev/{}/dcr-{}-images/data-clean-room-monitor".format(gcp_project_id, deploy_env)
registry_jupyter_image = "us-docker.pkg.dev/{}/dcr-{}-images/scipy-notebook-with-dcr".format(gcp_project_id, deploy_env)
registry_user_base_image = "us-docker.pkg.dev/{}/dcr-{}-user-images/data-clean-room-base".format(gcp_project_id, deploy_env)
