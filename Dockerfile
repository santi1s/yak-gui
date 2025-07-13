# We need to ruby image for helm export tasks
ARG BASE_VERSION=3.4.4-bookworm-20250526_220351@sha256:03df987e4572b23fae2782521b9e7b3c8a7a00c5d5dd812ffba2a76f2432927b
ARG ROOT_IMAGE=580698825394.dkr.ecr.eu-central-1.amazonaws.com/ruby
FROM ${ROOT_IMAGE}:${BASE_VERSION}

ARG TARGETARCH

ARG AWS_VAULT_VERSION=7.2.0
ARG AWS_VAULT_VERSION_SHA_AMD64="b92bcfc4a78aa8c547ae5920d196943268529c5dbc9c5aca80b797a18a5d0693"
ARG AWS_VAULT_VERSION_SHA_ARM64="176fa59b9e3981720361cd6ed87b5389741953c62ba0a6fa4f2db401f8bef735"
ENV AWS_CONFIG_FILE=
ENV AWS_VAULT_FILE_PASSPHRASE=
ENV AWS_VAULT_BACKEND=file

RUN curl -sfLo /usr/local/bin/aws-vault "https://github.com/99designs/aws-vault/releases/download/v${AWS_VAULT_VERSION}/aws-vault-linux-${TARGETARCH}" \
    && if [ "${TARGETARCH}" != "arm64" ]; then echo "${AWS_VAULT_VERSION_SHA_AMD64} /usr/local/bin/aws-vault" | sha256sum -c -; else echo "${AWS_VAULT_VERSION_SHA_ARM64} /usr/local/bin/aws-vault" | sha256sum -c -; fi

ARG HELM_VERSION=3.16.1
ARG HELM_VERSION_SHA_AMD64="e57e826410269d72be3113333dbfaac0d8dfdd1b0cc4e9cb08bdf97722731ca9"
ARG HELM_VERSION_SHA_ARM64="780b5b86f0db5546769b3e9f0204713bbdd2f6696dfdaac122fbe7f2f31541d2"

RUN curl -sfLo /tmp/helm.tar.gz "https://get.helm.sh/helm-v${HELM_VERSION}-linux-${TARGETARCH}.tar.gz" \
    && if [ "${TARGETARCH}" != "arm64" ]; then echo "${HELM_VERSION_SHA_AMD64} /tmp/helm.tar.gz" | sha256sum -c -; else echo "${HELM_VERSION_SHA_ARM64} /tmp/helm.tar.gz" | sha256sum -c -; fi \
    && tar xvzf /tmp/helm.tar.gz \
    && mv linux-${TARGETARCH}/helm /usr/local/bin/helm \
    && rm -rf linux-${TARGETARCH} /tmp/helm.tar.gz

RUN chmod +x /usr/local/bin/aws-vault

COPY docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

COPY yak /usr/local/bin

RUN yak --help \
    && helm version \
    && aws-vault --version

ENTRYPOINT ["/entrypoint.sh"]
