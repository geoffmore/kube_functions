#!/bin/bash
IMAGE=$1

# User path is $["Config"]["User"]
# User=0 is very bad
if [ -n ${IMAGE} ]; then
  docker inspect ${IMAGE}
  # There should be some filtering here
else 
  echo "No image specified"
fi
