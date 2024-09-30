# NOTE: this is a hacky way to load the configs, because it's not actually a .bzl file
# In the future, we need to remove the separate env file and rely on a single config file.
load("//:env.bzl", "project_id", "env")

# reassign external variables to bzl variables here.
gcp_project_id = project_id
environment = env

# artifact registries
registry_api_image = "us-docker.pkg.dev/{}/dcr-{}-images/data-clean-room-api".format(gcp_project_id, environment)
registry_monitor_image = "us-docker.pkg.dev/{}/dcr-{}-images/data-clean-room-monitor".format(gcp_project_id, environment)
registry_user_base_image = "us-docker.pkg.dev/{}/dcr-{}-user-images/data-clean-room-base".format(gcp_project_id, environment)