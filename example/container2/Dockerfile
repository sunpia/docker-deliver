FROM docker.io/library/container_base:latest AS base

WORKDIR /app
COPY environment.yaml .

RUN conda env update -n env --file environment.yaml \
    && conda clean -a -y


# Activate environment by default
SHELL ["conda", "run", "-p", "/opt/conda/envs/env", "/bin/bash", "-c"]

COPY hello_world.py /app

CMD ["conda", "run", "-p", "/opt/conda/envs/env", "python3", "/app/hello_world.py"]
