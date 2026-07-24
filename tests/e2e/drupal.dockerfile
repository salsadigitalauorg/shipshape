# Drupal + drush + shipshape image for the build-tagged (e2e_docker) e2e suite.
#
# Builds shipshape from the repo source and installs a Drupal codebase with
# drush and the PHPStan tooling the examples reference. Used by docker_test.go
# (//go:build e2e_docker) which installs a site (drush site:install) against a
# MariaDB container and runs the drush-backed examples inside this container.
#
# Build context must be the repository root.
FROM uselagoon/php-8.4-cli-drupal

ARG DRUPAL_VERSION=^11.0
ARG DRUSH_VERSION=^13.0
ARG PHPSTAN_EXTENSION_INSTALLER_VERSION=^1.3
ARG PHPSTAN_DISALLOWED_CALLS_VERSION=^4.6

RUN set -ex; \
    apk add --no-cache go; \
    rm -rf ~/.drush; \
    composer create-project drupal/recommended-project:${DRUPAL_VERSION} /app; \
    cd /app && composer config allow-plugins.phpstan/extension-installer true; \
    composer require \
        drush/drush:${DRUSH_VERSION} \
        phpstan/extension-installer:${PHPSTAN_EXTENSION_INSTALLER_VERSION} \
        spaze/phpstan-disallowed-calls:${PHPSTAN_DISALLOWED_CALLS_VERSION}; \
    ln -sf /app/vendor/bin/drush /usr/local/bin/drush

COPY tests/e2e/.docker/phpstan.neon /app/phpstan.neon

# Build shipshape from source and place it on PATH.
COPY . /shipshape
RUN set -ex; \
    cd /shipshape; \
    go generate ./...; \
    go build -o /usr/local/bin/shipshape .
