#!/bin/bash
set -e
readonly example_dir=$(dirname "$(dirname "$(readlink -f "$0")")")

dockerdeliver:benchmark:base() {
    # Print message in orange color
    echo -e "\e[38;5;208mStarting export of docker-deliver images...\e[0m"
    docker-deliver save \
        -f "$example_dir/docker-compose.base.yaml" \
        -f "$example_dir/docker-compose.extend.yaml" \
        --tag latest \
        -o tmp > /dev/null
    if [ -d tmp ]; then
            size=$(du -sh tmp | cut -f1)
            # Print folder size in blue color
            echo -e "\e[34mFolder size of tmp: $size\e[0m"
            rm -rf tmp
    fi
    echo "Export completed."
}

dockerdeliver:benchmark:image() {
    # Print message in orange color
    echo -e "\e[38;5;208mStarting export of Docker images...\e[0m"
    images=("container0" "container1" "container2" "container_base")
    imageTag="latest"
    for image in "${images[@]}"; do
        image="$image:$imageTag"
        filename="$image.tar"
        echo "Exporting $image to $filename..."
        docker save -o "$filename" "$image" > /dev/null
        if [ -f "$filename" ]; then
            filesize=$(du -h "$filename" | cut -f1)
            # Print file size in blue color
            echo -e "\e[34mFile size of $filename: $filesize\e[0m"
            rm "$filename"
            echo "Removed $filename"
        fi
    done
    echo "Export completed."
}

dockerdeliver:benchmark:conda() {
    # Print message in orange color
    echo -e "\e[38;5;208mStarting export of Conda environments...\e[0m"
    envs=("container0" "container1" "container2" "container_base")
    for env in "${envs[@]}"; do
        requirements_file="$example_dir/$env/requirements.yaml"
        conda env create -q -y -n $env -f $requirements_file > /dev/null
        echo "Packing Conda environment: $env"
        conda pack -q -n "$env" -o "${env}.tar" > /dev/null
        if [ -f "${env}.tar" ]; then
            filesize=$(du -h "${env}.tar" | cut -f1)
            # Print file size in blue color
            echo -e "\e[34mFile size of ${env}.tar: $filesize\e[0m"
            rm "${env}.tar"
            echo "Removed ${env}.tar"
        fi
        conda remove -y -n "$env" --all -q > /dev/null
    done
    echo "Conda environment export completed."
}

dockerdeliver::benchmark::run() {
    dockerdeliver:benchmark:base
    dockerdeliver:benchmark:image
    dockerdeliver:benchmark:conda
}

dockerdeliver::benchmark::run