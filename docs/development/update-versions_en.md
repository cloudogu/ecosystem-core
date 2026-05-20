## Update Components

To automatically update the component versions in `values.yaml` to the latest available versions,
the make target `make update-ecosystem-versions` can be used.

The target checks the corresponding cloudogu repositories for each component and writes the latest version into the YAML file
if the version has changed.

The log output of the target also contains the correct commit message including the component name, old version, and new version.

Some components are located in repositories whose names do not match the component name (e.g. CRD components are located in lib repositories).
For these cases, the `repo-mapping.txt` file can be adjusted.

A GIT_TOKEN is required to access Git via the API.
It can either be added to the .env file or passed directly to the make target:

`GIT_TOKEN=1234567890 make update-ecosystem-versions`