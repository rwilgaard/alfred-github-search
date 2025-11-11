#!/usr/bin/env bash

cat << EOB
{"items": [
    {
        "title": "Press ↩ for details",
        "subtitle": "$error_msg",
        "icon": {
            "path": "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertStopIcon.icns",
        },
    },
    {
        "title": "Clear credentials to reauthenticate",
        "arg": "clearauth",
    },
]}
EOB
