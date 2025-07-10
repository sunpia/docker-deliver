FROM continuumio/miniconda3:25.3.1-1 AS base

WORKDIR /app

USER root 
RUN apt-get update && apt-get install -y rsync

COPY requirements.yaml .

# Create environment and install Python packages
RUN conda env create -y -n env -f requirements.yaml && conda clean -ya
