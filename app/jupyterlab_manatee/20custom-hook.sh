#!/bin/bash

# this hook is executed before notebook starts.
pip install /manatee/jupyterlab_manatee-0.0.0-py3-none-any.whl
jupyter labextension disable @jupyterlab/docmanager-extension:download
jupyter labextension disable @jupyterlab/filebrowser-extension:download