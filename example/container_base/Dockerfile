FROM continuumio/miniconda3:25.3.1-1 AS base

WORKDIR /app

USER root 
RUN apt-get update && apt-get install -y rsync

COPY environment.yaml .

# Create environment and install Python packages
RUN conda env create -y -n env -f environment.yaml && conda clean -a -y
