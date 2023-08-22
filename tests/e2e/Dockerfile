ARG SHIPSHAPE_VERSION

FROM ghcr.io/salsadigitalauorg/shipshape:${SHIPSHAPE_VERSION:-latest} AS shipshape

FROM uselagoon/php-8.1-cli-drupal

ARG DRUPAL_VERSION=^9.5.4
ARG DRUSH_VERSION=^10.0
ARG PHPSTAN_EXTENSION_INSTALLER_VERSION=^1.3
ARG PHPSTAN_DISALLOWED_CALLS_VERSION=^2.3.0

COPY --from=shipshape /usr/local/bin/shipshape /usr/local/bin/shipshape

RUN set -ex; \
    # Add golang for when we want to compile shipshape.
    apk add --no-cache go; \
    \
    # Install drupal & related dependencies.
    rm -rf ~/.drush; \
    composer create-project drupal/recommended-project:${DRUPAL_VERSION} /app; \
    cd /app && composer config allow-plugins.phpstan/extension-installer true; \
    composer require \
        drush/drush:${DRUSH_VERSION} \
        phpstan/extension-installer:${PHPSTAN_EXTENSION_INSTALLER_VERSION} \
        spaze/phpstan-disallowed-calls:${PHPSTAN_DISALLOWED_CALLS_VERSION}; \
    \
    # Install venom.
    VENOM_LATEST=$(curl -Ls -o /dev/null \
        https://github.com/ovh/venom/releases/latest -w %{url_effective}); \
    VENOM_LATEST_VERSION=${VENOM_LATEST##*/}; \
    [ "$(uname -m)" == "x86_64" ] && ARCH=amd64 || ARCH=arm64; \
    curl -LO https://github.com/ovh/venom/releases/download/${VENOM_LATEST_VERSION}/venom.linux-${ARCH}; \
    chmod +x venom.linux-${ARCH}; \
    mv venom.linux-${ARCH} /usr/local/bin/venom;

COPY tests/e2e/.docker/phpstan.neon /app/phpstan.neon

COPY . /shipshape

RUN set -ex; \
    # Install shipshape.
    [ ! -z "$SHIPSHAPE_VERSION" ] && exit 0; \
    ls -l /shipshape; \
    cd /shipshape && go build -o /usr/local/bin/shipshape;
