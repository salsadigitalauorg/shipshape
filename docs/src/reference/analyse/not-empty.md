# not:empty

The `not:empty` analyser checks if a map contains any values. This is useful for ensuring configuration sections are populated.

## Configuration

No additional configuration is required beyond the common analyser fields.

<Content :page-key="$site.pages.find(p => p.path === '/reference/common/analyse.html').key"/>

## Supported Input Formats

- `FormatMapNestedString`: Checks if the nested map contains any values

## Example Usage

