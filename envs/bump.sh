#!/bin/bash -e
# This scripts updates a .json file like this:
#
# $ SERVICE_NAME=service_name IMAGE_TAG=99zz ./bump.sh file.json
#   { "service_name": "abc" } -> { "service_name": "99zz" }
#
# $ SERVICE_NAME=another IMAGE_TAG=bbb ./bump.sh file.json
#   { "service_name": "abc" } -> { "service_name": "abc", "another": "bbb" }

usage() {
  echo "Usage: SERVICE_NAME=svc IMAGE_TAG=123abc FILENAME=file.json $0"
  exit 1
}

if [ -z "${FILENAME}" ] || [ -z "${SERVICE_NAME}" ]; then
  usage
fi

if ! test -f "${FILENAME}"; then
  # file does not exist, init an empty json
  echo '{}' > "${FILENAME}"
fi

# update filename json
tmpfile=$(mktemp /tmp/bump-services.XXXXXX)
cat ${FILENAME} | jq -r ".\"${SERVICE_NAME}\"=\"${IMAGE_TAG}\"" > ${tmpfile}
mv ${tmpfile} ${FILENAME}

# print updated file
cat ${FILENAME}
