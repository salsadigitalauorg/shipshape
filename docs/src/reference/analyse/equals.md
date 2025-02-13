# equals

The `equals` analyser checks if a value equals a given string. For map inputs, it checks if the value at a specified key equals the given string.

## Configuration

| Field   | Type   | Required | Description                                      |
| ------- | ------ | -------- | ------------------------------------------------ |
| value   | string | Yes      | The string value to compare against              |
| key     | string | No       | For map inputs, the key whose value to check     |

<Content :page-key="$site.pages.find(p => p.path === '/reference/common/analyse.html').key"/>

## Supported Input Formats

- `FormatString`: Checks if the string value equals the configured value
- `FormatMapString`: Checks if the value at the specified key equals the configured value

## Example Usage

