#!/bin/bash
logger "Synctime START"
TIME_API_URL="${1:-http://worldtimeapi.org/api/timezone/Europe/Sofia}"
DATETIME=$(curl -s ${TIME_API_URL} | jq -r .datetime)
[[ ! -z "${DATETIME}" ]] && sudo date -s "${DATETIME}" && logger "Synctime OK: ${DATETIME}" || logger "Synctime FAIL"
