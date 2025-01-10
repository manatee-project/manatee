load("@gazelle//:def.bzl", "gazelle")
load("@rules_multirun//:defs.bzl", "multirun")
load("@rules_oci//oci:defs.bzl", "oci_push")
load("//:env.bzl", "env", "project_id", "region", "zone")

# gazelle:prefix github.com/manatee-project/manatee
gazelle(name = "gazelle")

REPOS = {
    "api": "us-docker.pkg.dev/{}/dcr-{}-$$namespace-images/manatee-api".format(project_id, env),
    "reconciler": "us-docker.pkg.dev/{}/dcr-{}-$$namespace-images/manatee-reconciler".format(project_id, env),
    "jupyterlab_manatee": "us-docker.pkg.dev/{}/dcr-{}-$$namespace-images/scipy-notebook-with-dcr".format(project_id, env),
    "executor": "us-docker.pkg.dev/{}/dcr-{}-user-images/manatee-executor-base".format(project_id, env),
}

[
    genrule(
        name = "{}_repo".format(k),
        outs = ["{}_repo.txt".format(k)],
        cmd = "echo '{}' | envsubst > $@".format(v),
    )
    for (k, v) in REPOS.items()
]

[
    oci_push(
        name = "push_{}_image".format(k),
        image = "//app/{}:image".format(k),
        remote_tags = ["latest"],
        repository_file = ":{}_repo".format(k),
    )
    for k in REPOS.keys()
]

multirun(
    name = "push_all_images",
    commands = [
        "push_{}_image".format(k)
        for k in REPOS.keys()
    ],
    jobs = 0,
)

multirun(
    name = "load_all_images",
    commands = [
        "//app/{}:load_image".format(k)
        for k in REPOS.keys()
    ],
)
