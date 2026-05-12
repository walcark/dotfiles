#!/usr/bin/env bash

export MOPSMAP_DATASET_PATH="$HOME/dev/third-party/mopsmap/optical_dataset"

# PostgreSQL
export DB_USER=kevin
export DEFAULT_DB_HOST=localhost
export DEFAULT_DB_PORT=5432

if [ -f "$HOME/.postgresql" ]; then
    source "$HOME/.postgresql"
fi
