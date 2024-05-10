#!/bin/bash

# Set the default branch to 'master'
branch=${1:-master}

# Repository URL
repo_url="https://github.com/vechain/thor.git"

# Directory where the repository will be cloned
tmp_dir="$(pwd)/tmp"
clone_dir="${tmp_dir}/thor_repo"
mkdir -p $clone_dir
echo "Cloning the branch '$branch' of repository '$repo_url'..."

# Clone the specified branch
git clone --branch $branch $repo_url $clone_dir

# Check if the clone was successful
if [ $? -ne 0 ]; then
    echo "Failed to clone the repository."
    exit 1
fi

echo "Repository cloned successfully."

# Change to the repository directory
cd $clone_dir

# Build the thor binary
echo "Building the thor binary..."
make thor

# Check if the build was successful
if [ $? -ne 0 ]; then
    echo "Failed to build the thor binary."
    exit 1
fi

echo "thor binary built successfully."
cp "$clone_dir/bin/thor" ../..
