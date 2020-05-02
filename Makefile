#############################
# VARIABLES
#############################
START 										:= $(shell printf "\033[34;1m")
END 											:= $(shell printf "\033[0m")
SHELL 										:= /bin/bash
BUILD_SCRIPTS 						:= __build__/scripts
ENV   										?= dev
LOG_LEVEL									?= warn

#############################
# CUSTOM FUNCTIONS
#############################
define header
  $(info $(START)▶▶▶ $(1)$(END))
endef

#############################
# PHONIES
#############################
.PHONY: all before_deploy clean

#############################
# TARGETS
#############################
all: before-deploy deploy

deploy: before-deploy
	$(call header,DEPLOYING...)
	$(SHELL) $(BUILD_SCRIPTS)/skaffold.sh $(LOG_LEVEL) $(DEV_CRONJOB)

before-deploy:
	$(call header,BEFORE DEPLOY...)
	$(SHELL) $(BUILD_SCRIPTS)/before_deploy.sh

# before-build:
# 	$(call header,BEFORE BUILD...)
# 	$(SHELL) $(BUILD_SCRIPTS)/before_build.sh
