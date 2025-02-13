# regex:not-match

The `regex:not-match` analyser checks if a value does NOT match a given regular expression pattern. This is useful for ensuring certain patterns are absent from your configuration.

## Configuration

| Field   | Type   | Required | Description                                          |
| ------- | ------ | -------- | ---------------------------------------------------- |
| pattern | string | Yes      | The regular expression pattern that should NOT match |


<Content :page-key="$site.pages.find(p => p.path === '/reference/common/analyse.html').key"/>

## Supported Input Formats

- `FormatNil`: No validation performed
- `FormatMapNestedString`: Checks all nested string values against the pattern
- `FormatString`: Checks the string value against the pattern

## Example Usage
