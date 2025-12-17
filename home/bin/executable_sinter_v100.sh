#!/usr/bin/env bash

set -euo pipefail

# ---------- Check arguments ----------
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <hours (1-99)>"
    exit 1
fi

HOURS="$1"

# Check integer
if ! [[ "$HOURS" =~ ^[0-9]+$ ]]; then
    echo "Error: hours must be an integer."
    exit 1
fi

# Check range
if [ "$HOURS" -lt 1 ] || [ "$HOURS" -gt 99 ]; then
    echo "Error: hours must be between 1 and 99."
    exit 1
fi

# Format time HH:00:00 (zero-padded)
TIME_FMT=$(printf "%02d:00:00" "$HOURS")

echo "Launching sinter job for ${TIME_FMT} ..."

# ---------- Run command ----------

unset SLURM_JOB_ID 
srun \
    -A cesbio \
    -N 1 \
    -n 5 \
    --mem-per-cpu=12000M \
    --partition=gpu_std \
    --qos=gpu_all \
    --gpus=v100_32g:1 \
    --time="$TIME_FMT" \
    --x11 \
    --pty /bin/bash
