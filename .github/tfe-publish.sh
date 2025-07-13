#!/bin/bash -x

PROVIDER_NAME=$1
PROVIDER_VERSION=$2

# Create provider in TFE registry
echo "Create provider ${PROVIDER_NAME} in TFE registry"
curl \
  --header "Authorization: Bearer ${TFE_TOKEN}" \
  --header "Content-Type: application/vnd.api+json" \
  --request POST \
  --data "{ \
  \"data\": { \
    \"type\": \"registry-providers\", \
    \"attributes\": { \
      \"name\": \"${PROVIDER_NAME}\", \
      \"namespace\": \"${TFE_NAMESPACE}\", \
      \"registry-name\": \"private\" \
    } \
  } \
}" \
  https://${TFE_HOSTNAME}/api/v2/organizations/doctolib/registry-providers

# Create version in TFE registry
echo "Create provider version ${PROVIDER_VERSION} in TFE registry"
sha256sum_urls=$(curl \
  --header "Authorization: Bearer ${TFE_TOKEN}" \
  --header "Content-Type: application/vnd.api+json" \
  --request POST \
  --data "{ \
  \"data\": { \
    \"type\": \"registry-provider-versions\", \
    \"attributes\": { \
      \"version\": \"${PROVIDER_VERSION}\", \
      \"key-id\": \"${GPG_KEY_ID}\", \
      \"protocols\": [\"5.0\"] \
    } \
  } \
}" \
  https://${TFE_HOSTNAME}/api/v2/organizations/${TFE_NAMESPACE}/registry-providers/private/${TFE_NAMESPACE}/${PROVIDER_NAME}/versions
)

# Upload sha256sums and sha256sums.sig files
upload_url=$(echo $sha256sum_urls | jq -r .data.links.\"shasums-upload\")
sig_upload_url=$(echo $sha256sum_urls | jq -r .data.links.\"shasums-sig-upload\")
echo "Upload SHA256SUMS file"
curl \
  -T terraform-provider-${PROVIDER_NAME}_${PROVIDER_VERSION}_SHA256SUMS \
  ${upload_url}

echo "Upload SHA256SUMS.sig file"
curl \
  -T terraform-provider-${PROVIDER_NAME}_${PROVIDER_VERSION}_SHA256SUMS.sig \
  ${sig_upload_url}

# Create provider version platforms and upload the binary
  echo "Create provider version platform for ${BINARY_OS}_${BINARY_ARCH}"
  provider_platform=$(curl \
    --header "Authorization: Bearer ${TFE_TOKEN}" \
    --header "Content-Type: application/vnd.api+json" \
    --request POST \
    --data "{ \
    \"data\": {
      \"type\": \"registry-provider-version-platforms\",
      \"attributes\": {
        \"os\": \"${BINARY_OS}\",
        \"arch\": \"${BINARY_ARCH}\",
        \"shasum\": \"$(cat terraform-provider-${PROVIDER_NAME}_${PROVIDER_VERSION}_SHA256SUMS | grep ${BINARY_OS}_${BINARY_ARCH} | cut -d' ' -f1)\",
        \"filename\": \"terraform-provider-${PROVIDER_NAME}_${PROVIDER_VERSION}_${BINARY_OS}_${BINARY_ARCH}.zip\"
      }
    }
  }
  " \
    https://${TFE_HOSTNAME}/api/v2/organizations/${TFE_NAMESPACE}/registry-providers/private/${TFE_NAMESPACE}/${PROVIDER_NAME}/versions/${PROVIDER_VERSION}/platforms
  )

  echo "Upload binary for ${BINARY_OS}_${BINARY_ARCH}"
  platform_upload_url=$(echo ${provider_platform} | jq -r .data.links.\"provider-binary-upload\")

  curl -T terraform-provider-${PROVIDER_NAME}_${PROVIDER_VERSION}_${BINARY_OS}_${BINARY_ARCH}.zip \
    ${platform_upload_url}
