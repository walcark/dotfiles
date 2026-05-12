#!/usr/bin/env bash

# --- Modules ---
module load cuda/11.7.0
module load latex

# --- CUDA ---
export CUDA_HOME=/usr/local/cuda
export PATH="$CUDA_HOME/bin:$PATH"
export LD_LIBRARY_PATH="$CUDA_HOME/lib64:$LD_LIBRARY_PATH"

# --- SLURM helpers ---
alias sinter='unset SLURM_JOB_ID ; srun'
