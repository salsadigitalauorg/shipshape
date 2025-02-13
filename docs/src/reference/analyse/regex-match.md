# regex:match

The `regex:match` analyser checks if a value matches a given regular expression pattern. This is useful for validating that configuration values follow certain patterns.

## Configuration

| Field   | Type   | Required | Description                                      |
| ------- | ------ | -------- | ------------------------------------------------ |
| pattern | string | Yes      | The regular expression pattern that should match |


<Content :page-key="$site.pages.find(p => p.path === '/reference/common/analyse.html').key"/>

## Supported Input Formats

- `FormatNil`: No validation performed
- `FormatMapNestedString`: Checks all nested string values against the pattern
- `FormatString`: Checks the string value against the pattern

## Example Usage

