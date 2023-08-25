# Docker images

For building and pushing the docker base images:

1. Checkout branch called `base-image`.
1. Make needed modifications to Dockerfiles.
1. Push branch to remote, GitHub actions will take care of building.
1. Craft PR and merge changes as usual.

Note: base images are currently lacking proper versioning, only tag "latest" is used.
