# AWS s3 tool

This module contains helper tool for copying files from AWS s3 buckets.

There are also helper scripts for building base images (to AWS environment) for findy services.

## Usage

Define environment variables `AWS_DEFAULT_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`.

Run

```sh
go run . <bucket_name> <subfolder_in_bucket> <local_target_folder>
```

## Docker images

For building and pushing the docker base images:
1. Checkout branch called `base-image`.
1. Make needed modifications to Dockerfiles.
1. Push branch to remote, GitHub actions will take care of building.
1. Craft PR and merge changes as usual.

Note: base images are currently lacking proper versioning, only tag "latest" is used.
