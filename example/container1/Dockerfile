FROM mambaorg/micromamba:debian12-slim AS base

WORKDIR /app

# Copy environment.yaml
COPY environment.yaml .

# Create environment and install Python packages
RUN micromamba create -y -n env -f environment.yaml \
    && micromamba clean -a -y

SHELL ["micromamba", "run", "-n", "env", "/bin/bash", "-c"]

FROM base

COPY environment-prod.yaml .

RUN micromamba env update -n env -f environment-prod.yaml \
    && micromamba clean -a -y

