IMAGE_NAME_ALPINE:=findy-base:alpine-3.13
FULL_IMAGE_NAME_ALPINE:=ghcr.io/findy-network/$(IMAGE_NAME_ALPINE)
IMAGE_NAME_INDY_UBUNTU:=findy-base:indy-1.16.ubuntu-18.04
FULL_IMAGE_NAME_INDY_UBUNTU:=ghcr.io/findy-network/$(IMAGE_NAME_INDY_UBUNTU)

build:
	docker build -f Dockerfile.alpine -t $(IMAGE_NAME_ALPINE) .
	docker tag $(IMAGE_NAME_ALPINE) $(FULL_IMAGE_NAME_ALPINE)
	docker build -f Dockerfile.indy.ubuntu -t $(IMAGE_NAME_INDY_UBUNTU) .
	docker tag $(IMAGE_NAME_INDY_UBUNTU) $(FULL_IMAGE_NAME_INDY_UBUNTU)

push: build
	@[ "${CR_PAT}" ] || ( echo "ERROR: CR_PAT is not set, create GitHub access token with packages scope"; exit 1 )
	@[ "${CR_USER}" ] || ( echo "ERROR: CR_USER is not set, define GitHub username"; exit 1 )
	echo $(CR_PAT) | docker login ghcr.io -u $(CR_USER) --password-stdin
	docker push $(FULL_IMAGE_NAME_ALPINE)
	docker push $(FULL_IMAGE_NAME_INDY_UBUNTU)
	docker logout
