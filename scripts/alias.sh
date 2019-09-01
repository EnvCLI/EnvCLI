#!/usr/bin/env bash

# debug mode
DEBUG=${DEBUG:-false}
if [ "$DEBUG" == "true" ]; then
    set -x
fi

# find out which alias is called
aliasFor=$(echo $(basename $(readlink -nf $0)) | cut -f 1 -d '.')

# quote all args
for arg in "$@"; do 
    arg="${arg//\\/\\\\}"
    allargs="$allargs \"${arg//\"/\\\"}\""
done

# call envcli for the alias and pass all arguments
ENVCLI_DEBUG=${ENVCLI_DEBUG:-false}
if [ "$ENVCLI_DEBUG" == "true" ]; then
    eval envcli --loglevel=debug run $aliasFor $allargs
else
    eval envcli run $aliasFor $allargs
fi
