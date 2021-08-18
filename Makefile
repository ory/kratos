include .env
export $(shell sed 's/=.*//' .env)

CONTAINER_NAME := kratos
ECR_DIRECTORY  := base.cas/api/$(CONTAINER_NAME)
ECR_DOMAIN     := 140484324944.dkr.ecr.ap-northeast-1.amazonaws.com
ECR_PATH       := $(ECR_DOMAIN)/$(ECR_DIRECTORY)
LOCAL_TAG_BASE := aws-emcloud-$(CONTAINER_NAME)-base

.PHONY: branch-var-check
.SILENT: branch-var-check
branch-var-check:
	if [ -z "$(KRATOS_IMAGE_TAG_BRANCH_NAME)" ]; then \
		echo "Error: environment variable 'KRATOS_IMAGE_TAG_BRANCH_NAME' is not set. Check your .env file!"; \
		exit 1; \
	fi

.PHONY: version-var-check
.SILENT: version-var-check
version-var-check:
	if [ -z "$(KRATOS_IMAGE_TAG_VERSION)" ]; then \
		echo "Error: environment variable 'KRATOS_IMAGE_TAG_VERSION' is not set. Check your .env file!"; \
		exit 1; \
	fi

auth:
	@printf "\n------ Authenticating with AWS ------ \n\n"
	aws ecr get-login-password --region ap-northeast-1 | docker login --username AWS --password-stdin $(ECR_DOMAIN)

build:
	@printf "\n------ Building local image `tput setaf 3`$(LOCAL_TAG)`tput sgr0` ------ \n\n"
	docker build -f .docker/Dockerfile-build -t $(LOCAL_TAG) .

# get-local-hash:
# 	$(eval LOCAL_HASH := $(shell docker images --no-trunc --quiet $(LOCAL_TAG) | awk -F: '{ print substr($$2,1,12) }'))

# aws-tag: get-local-hash
aws-tag:
	@printf "\n------ Preparing image `tput setaf 3`$(ECR_DIRECTORY):$(TAG)`tput sgr0` for AWS ------ \n\n"
	docker tag $(LOCAL_TAG) $(ECR_PATH):$(TAG)
# docker tag $(LOCAL_TAG) $(ECR_PATH):$(LOCAL_HASH)

# aws-push: get-local-hash
aws-push:
	@printf "\n------ Pushing image `tput setaf 3`$(ECR_DIRECTORY):$(TAG)`tput sgr0` to AWS ------ \n\n"
	docker push $(ECR_PATH):$(TAG)
# docker push $(ECR_PATH):$(LOCAL_HASH)

build-publish: LOCAL_TAG="$(LOCAL_TAG_BASE):$(TAG)"
build-publish: build auth aws-tag aws-push alert-success

.SILENT: branch-safety-check
branch-safety-check:
	@printf "The following image tag will be used: `tput setaf 3`$(KRATOS_IMAGE_TAG_BRANCH_NAME)`tput sgr0`\n\
	This version will be UPLOADED to AWS and may override existing versions of the same name if any.\n\n"; \
	read -p "Type y to proceed: " consent; \
	if [[ $$consent == "y" ]]; then \
		echo "Proceeding..."; \
	else \
		echo "Aborting..."; \
		exit 0; \
	fi

.SILENT: version-safety-check
version-safety-check:
	@printf "`tput smso`SAFETY CHECK`tput sgr0`: please retype the version number to publish (it should match what is in your .env file).\n\
	This version will be `tput smul`UPLOADED to AWS and may override existing versions`tput sgr0` if the number is the same.\n\n"; \
	read -p "Version number to publish: " version; \
	if [[ $$version == "$(KRATOS_IMAGE_TAG_VERSION)" ]]; then \
		echo "Proceeding..."; \
	else \
		echo "Sorry, this does not match the value in your .env file."; \
		exit 1; \
	fi

.SILENT: alert-success
alert-success:
	@printf "\n`tput setaf 2`The build and upload to AWS succeeded`tput sgr0`\n"

.SILENT: warn-envfile
warn-envfile:
	@printf "\n`tput smso`IMPORTANT`tput sgr0`: `tput smul`remove the tag version from your .env file`tput sgr0` just to be extra safe.\nThe next upload will likely use a different version number and you don't want to risk overriding this one now do you?\n"

# build and publish only the branch tag to AWS
build-publish-branch: branch-var-check
build-publish-branch: TAG:=$(KRATOS_IMAGE_TAG_BRANCH_NAME)
build-publish-branch: branch-safety-check
build-publish-branch: build-publish

# build and publish only the version tag to AWS
build-publish-version: version-var-check
build-publish-version: TAG:=$(KRATOS_IMAGE_TAG_VERSION)
build-publish-version: version-safety-check
build-publish-version: build-publish
build-publish-version: warn-envfile
