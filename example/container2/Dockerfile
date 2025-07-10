FROM container_base AS base

WORKDIR /app
COPY requirements.yaml .

RUN conda env update -n env --file requirements.yaml \
    && conda clean -a -y


# Activate environment by default
SHELL ["conda", "run", "-p", "/opt/conda/envs/env", "/bin/bash", "-c"]

COPY hello_world.py /app

CMD ["conda", "run", "-p", "/opt/conda/envs/env", "python3", "/app/hello_world.py"]
