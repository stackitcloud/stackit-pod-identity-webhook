#!/usr/bin/env bash

set -euo pipefail

# This script packages and pushes Helm charts to an OCI registry.
# It expects the image full tag and the chart name as arguments.
#
# Usage: ./hack/push-chart.sh <image-tag> <chart-name>
# Example: ./hack/push-chart.sh ghcr.io/stackitcloud/stackit-pod-identity-webhook:v0.1.0 stackit-pod-identity-webhook

IMAGE_TAG_FULL="${1:-}"
CHART_NAME="${2:-}"
HELM_ARTIFACTS_DIR="artifacts/charts"
HELM="go tool helm"

if [[ -z "$IMAGE_TAG_FULL" || -z "$CHART_NAME" ]]; then
    echo "Usage: $0 <image-tag> <chart-name>"
    echo "  <image-tag>: Full image tag (e.g., ghcr.io/org/repo:v1.0.0)"
    echo "  <chart-name>: Name of the chart (e.g., stackit-pod-identity-webhook)"
    exit 1
fi

if [[ "${PUSH:-false}" != "true" ]]; then
    echo "PUSH is not set to true. Skipping chart push."
    exit 0
fi

# Function to extract image repository from full tag
# Example: ghcr.io/org/repo:v1.0.0 -> ghcr.io/org/repo
# Example: ghcr.io/org/repo@sha256:abcd -> ghcr.io/org/repo
function get_image_repo() {
    echo "$IMAGE_TAG_FULL" | cut -d'@' -f1 | cut -d':' -f1
}

# Function to extract image tag (including digest if present)
# Example: ghcr.io/org/repo:v1.0.0 -> v1.0.0
# Example: ghcr.io/org/repo:v1.0.0@sha256:abcd -> v1.0.0@sha256:abcd
function get_image_tag_with_digest() {
    if [[ "$IMAGE_TAG_FULL" == *":"* ]]; then
        echo "$IMAGE_TAG_FULL" | cut -d':' -f2-
    else
        echo "latest"
    fi
}

# Function to extract a clean version (no digest, no -dirty)
# Example: v1.0.0@sha256:abcd -> v1.0.0
# Example: v1.0.0-dirty -> v1.0.0
function get_clean_version() {
    get_image_tag_with_digest | cut -d'@' -f1 | sed 's/-dirty//'
}

# Function to extract OCI registry/org (parent of the image repository)
# Example: ghcr.io/org/repo -> ghcr.io/org
function get_oci_base() {
    local repo
    repo=$(get_image_repo)
    if [[ "$repo" == *"/"* ]]; then
        echo "$repo" | rev | cut -d'/' -f2- | rev
    else
        echo "$repo"
    fi
}

function package_and_push_chart() {
    local chart_path="$1"
    local chart_id
    chart_id=$(basename "$chart_path")
    
    local chart_build_dir="${HELM_ARTIFACTS_DIR}/${chart_id}"
    
    echo "--- Processing chart: ${chart_id} ---"
    
    rm -rf "$chart_build_dir"
    mkdir -p "$chart_build_dir"
    cp -r "${chart_path}/." "$chart_build_dir"

    local image_repo
    image_repo=$(get_image_repo)
    local image_tag_full
    image_tag_full=$(get_image_tag_with_digest)

    echo "Updating values.yaml with image repository: ${image_repo} and tag: ${image_tag_full}"
    yq -i "
      .image.repository = \"${image_repo}\" | 
      .image.tag = \"${image_tag_full}\"
    " "$chart_build_dir/values.yaml"

    local chart_version
    chart_version=$(get_clean_version)
    
    echo "Packaging chart ${chart_id} with version ${chart_version}..."
    local helm_package_output
    if ! helm_package_output=$(${HELM} package "$chart_build_dir" --version "$chart_version" -d "$HELM_ARTIFACTS_DIR" 2>&1); then
        echo "Error: helm package failed for chart ${chart_id}."
        echo "$helm_package_output"
        exit 1
    fi

    local packaged_chart_file
    packaged_chart_file=$(echo "$helm_package_output" | sed -n 's/Successfully packaged chart \(into: \|and saved it to: \)\(.*\)/\2/p')

    if [[ -z "$packaged_chart_file" || ! -f "$packaged_chart_file" ]]; then
        echo "Error: Could not find packaged chart file for ${chart_id}."
        echo "$helm_package_output"
        exit 1
    fi

    local oci_base
    oci_base=$(get_oci_base)
    local push_url="oci://${oci_base}/charts"

    echo "Pushing ${packaged_chart_file} to ${push_url}..."
    local push_output
    if ! push_output=$(${HELM} push "$packaged_chart_file" "$push_url" 2>&1); then
        echo "Error: Failed to push chart ${chart_id}."
        echo "$push_output"
        exit 1
    fi

    echo "$push_output" | grep -E "Pushed:|Digest:"
}

# Main execution
rm -rf "$HELM_ARTIFACTS_DIR"
mkdir -p "$HELM_ARTIFACTS_DIR"

CHART_SOURCE_DIR="charts/${CHART_NAME}"

if [[ ! -d "$CHART_SOURCE_DIR" ]]; then
    echo "Error: Chart directory $CHART_SOURCE_DIR does not exist."
    exit 1
fi

# Package and push the main chart
package_and_push_chart "$CHART_SOURCE_DIR"
