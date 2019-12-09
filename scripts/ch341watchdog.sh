#!/bin/bash

NOW=$(date +%s)
LASTERR=$(dmesg --time-format iso  | grep 'usb 1-1: failed to send control message: -110'| tail -1 | awk -F ',' '{gsub("T", " "); print}')
LASTREBOOT=$(grep "USBWATCHDOG CH341 reboot" /var/log/syslog | tail -1 | awk '{print $NF}')
DUR=$((NOW - LASTREBOOT))

if [[ ! -z "${LASTERR}" ]]; then
    if (( $DUR < 600 )) ; then
        logger "USBWATCHDOG CH341 last reboot was too soon ${LASTREBOOT}"
        exit 0
    fi

    logger "USBWATCHDOG CH341 reboot ${NOW}"
    sudo reboot
fi
