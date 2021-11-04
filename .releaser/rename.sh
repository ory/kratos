#!/bin/bash

# workaround script as goreleaser doesnt support 'replacements' in builds section
# needed to adjust cyclonedx-gomod sbom files to match archive file names
# https://github.com/goreleaser/goreleaser/issues/2617
filename=$1
filename_adjusted=${filename//darwin/macos}
filename_adjusted=${filename_adjusted//386/32bit}
filename_adjusted=${filename_adjusted//amd64/64bit}
filename_adjusted=${filename_adjusted//arm_5/arm32v5}
filename_adjusted=${filename_adjusted//arm_6/arm32v6}
filename_adjusted=${filename_adjusted//arm_7/arm32v7}

if [ "$filename" != "$filename_adjusted" ]; then 
  echo "Renaming '$filename' to '$filename_adjusted' ..."
  mv "$filename" "$filename_adjusted" 
else 
  echo "Skipping file '$filename' ..."
fi

