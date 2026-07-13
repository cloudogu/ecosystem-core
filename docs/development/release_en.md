## Performing a Release

A release is started via the interactive git flow release workflow:

```bash
make ecosystem-core-release
```

Before the release, the component versions in `values.yaml` should be updated
(see [Update Components](update-versions_en.md)).

### Notes on the Release PR

- The component versions entered must match and be aligned with each other.
- If a new version of a component depends on another component that has not been released yet,
  a draft PR can be created first. Once the dependent component has been released, its new
  version is entered and only then is the PR submitted for review.

### Reviewing and Testing the Release PR

- Look at the release notes of the updated components and take them into account for testing.
- Test at least once with the default configuration to verify that everything works with it.
- Depending on the release notes of the updated components, perform additional tests
  (e.g. enable LOP-IdP or set other configurations).
