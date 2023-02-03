#!/bin/bash
set -ex

#############
# VARIABLES #
#############
HELM_ERR="changes found by 'make helm'. please run 'make helm' locally and update your PR"

#############
# FUNCTIONS #
#############
function validate(){
  git=$(git status --porcelain)
  for file in $git; do
    if [ -n "$file" ]; then
      echo "$1"
      echo "$git"
      git diff
      exit 1
    fi
  done
}

########
# MAIN #
########
case "$1" in
  helm)
    validate "$HELM_ERR"
  ;;
  *)
    echo $"Usage: $0 {helm}"
    exit 1
esac
