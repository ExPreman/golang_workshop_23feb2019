#!/bin/sh

set -eu
cd /

cmd=${1:-}
if [ -z "$cmd" ]; then
  echo "no command" >&2
  exit 1
fi

shift

case "$cmd" in
createuser|dbmigrate|server|worker)
  exec "/app/$cmd" "$@"
  ;;
*)
  echo "command $cmd is not available"
  ;;
esac
