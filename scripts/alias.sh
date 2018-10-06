#!/bin/bash

# find out which alias is called
aliasFor=$(echo $(basename $(readlink -nf $0)) | cut -f 1 -d '.')

# call envcli for the alias and pass all arguments
envcli run $aliasFor $@
