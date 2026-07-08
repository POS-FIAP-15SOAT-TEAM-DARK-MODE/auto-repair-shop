# Toolbox image: run the whole local pipeline with only Docker on the host.
# Bundles docker-cli, kind and kubectl (plus terraform for the AWS dry-run).
# Docker-outside-of-Docker: the host Docker socket is mounted in, so kind/build
# create sibling containers on the host daemon.
FROM alpine:3.20

# Provided automatically by BuildKit (amd64 / arm64).
ARG TARGETARCH=amd64
ARG KIND_VERSION=v0.24.0
ARG KUBECTL_VERSION=v1.31.2
ARG TERRAFORM_VERSION=1.9.8

RUN apk add --no-cache bash make curl git unzip docker-cli aws-cli

# kubectl
RUN curl -fsSLo /usr/local/bin/kubectl \
      "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${TARGETARCH}/kubectl" \
 && chmod +x /usr/local/bin/kubectl

# kind
RUN curl -fsSLo /usr/local/bin/kind \
      "https://kind.sigs.k8s.io/dl/${KIND_VERSION}/kind-linux-${TARGETARCH}" \
 && chmod +x /usr/local/bin/kind

# terraform (for the AWS dry-run targets)
RUN curl -fsSLo /tmp/tf.zip \
      "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_${TARGETARCH}.zip" \
 && unzip -d /usr/local/bin /tmp/tf.zip && rm /tmp/tf.zip

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

WORKDIR /workspace
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["bash"]
