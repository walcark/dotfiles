#!/usr/bin/env bash

if [ -x "$(command -v pixi)" ]; then
    eval "$(pixi completion --shell bash)"
fi
