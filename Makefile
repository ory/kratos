include .env
export $(shell sed 's/=.*//' .env)

CONTAINER_NAME := kratos
ECR_DIRECTORY_VERSIONS  := base.cas/api/$(CONTAINER_NAME)
ECR_DIRECTORY_BRANCHES  := base.cas/api/$(CONTAINER_NAME)-dev
ECR_DOMAIN     := 140484324944.dkr.ecr.ap-northeast-1.amazonaws.com
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

aws-tag:
	@printf "\n------ Preparing image `tput setaf 3`$(ECR_DIRECTORY):$(TAG)`tput sgr0` for AWS ------ \n\n"
	docker tag $(LOCAL_TAG) $(ECR_PATH):$(TAG)

aws-push:
	@printf "\n------ Pushing image `tput setaf 3`$(ECR_DIRECTORY):$(TAG)`tput sgr0` to AWS ------ \n\n"
	docker push $(ECR_PATH):$(TAG)

build-publish: LOCAL_TAG="$(LOCAL_TAG_BASE):$(TAG)"
build-publish: ECR_PATH=$(ECR_DOMAIN)/$(ECR_DIRECTORY)
build-publish: build auth aws-tag aws-push alert-success

.SILENT: branch-safety-check
branch-safety-check:
	@printf "The following image tag will be used: `tput setaf 3`$(KRATOS_IMAGE_TAG_BRANCH_NAME)`tput sgr0`\n\
	This branch image will be `tput bold`UPLOADED to AWS`tput sgr0` and `tput smul`may override existing versions`tput sgr0` of the same name if any.\n\n"; \
	read -p "Type y to proceed: " consent; \
	if [ $$consent == "y" ]; then \
		echo "Proceeding..."; \
	else \
		echo "Aborting..."; \
		exit 1; \
	fi

.SILENT: version-safety-check
version-safety-check:
	@printf "`tput smso`SAFETY CHECK`tput sgr0`: please retype the version number to publish (it should match what is in your .env file).\n\
	This version image will be `tput bold``tput smul`UPLOADED to AWS and will fail`tput sgr0` if it matches an existing version.\n\n"; \
	read -p "Version number to publish: " version; \
	if [ $$version == "$(KRATOS_IMAGE_TAG_VERSION)" ]; then \
		echo "Proceeding..."; \
	else \
		echo "Sorry, this does not match the value in your .env file."; \
		exit 1; \
	fi

.SILENT: alert-success
alert-success:
	@printf "\n`tput setaf 2`The build and upload to AWS succeeded`tput sgr0`\n"

# build and publish only the branch tag to AWS
build-publish-branch: branch-var-check
build-publish-branch: ECR_DIRECTORY:=$(ECR_DIRECTORY_BRANCHES)
build-publish-branch: TAG:=$(KRATOS_IMAGE_TAG_BRANCH_NAME)
build-publish-branch: branch-safety-check
build-publish-branch: build-publish

# build and publish only the version tag to AWS
build-publish-version: version-var-check
build-publish-version: ECR_DIRECTORY:=$(ECR_DIRECTORY_VERSIONS)
build-publish-version: TAG:=$(KRATOS_IMAGE_TAG_VERSION)
build-publish-version: version-safety-check
build-publish-version: build-publish
