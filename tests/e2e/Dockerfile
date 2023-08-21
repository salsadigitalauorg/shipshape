ARG SHIPSHAPE_VERSION=latest

FROM ghcr.io/salsadigitalauorg/shipshape:${SHIPSHAPE_VERSION} AS shipshape

FROM uselagoon/php-8.1-cli-drupal

ARG DRUPAL_VERSION=^9.5.4
ARG DRUSH_VERSION=^10.0

COPY --from=shipshape /usr/local/bin/shipshape /usr/local/bin/shipshape

RUN set -ex; \
    rm -rf ~/.drush; \
    composer create-project drupal/recommended-project:${DRUPAL_VERSION} /app; \
    cd /app && composer require drush/drush:${DRUSH_VERSION}; \
    \
    # Install venom.
    VENOM_LATEST=$(curl -Ls -o /dev/null \
        https://github.com/ovh/venom/releases/latest -w %{url_effective}); \
    VENOM_LATEST_VERSION=${VENOM_LATEST##*/}; \
    [[ "$(uname -m)" == "x86_64" ]] && ARCH=amd64 || ARCH=arm64; \
    curl -LO https://github.com/ovh/venom/releases/download/${VENOM_LATEST_VERSION}/venom.linux-${ARCH}; \
    chmod +x venom.linux-${ARCH}; \
    mv venom.linux-${ARCH} /usr/local/bin/venom;
