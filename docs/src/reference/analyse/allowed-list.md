# allowed:list

The `allowed:list` analyser checks if values in a list or map match against a list of allowed values. It can also enforce required values and flag deprecated values.

## Configuration

| Field         | Type     | Required | Description                                                |
| ------------- | -------- | -------- | ---------------------------------------------------------- |
| allowed       | []string | No       | List of allowed values                                     |
| required      | []string | No       | List of values that must be present                        |
| deprecated    | []string | No       | List of deprecated values to flag                          |
| exclude-keys  | []string | No       | For map inputs, keys to exclude from validation            |
| ignore        | []string | No       | List of values to ignore during validation                 |
| package-match | string   | No       | If set, treats values as packages and matches package names |

<Content :page-key="$site.pages.find(p => p.path === '/reference/common/analyse.html').key"/>

## Supported Input Formats

- `FormatListString`: Validates each string in the list
- `FormatMapString`: Validates each value in the map
- `FormatMapListString`: Validates each string in the lists contained in the map

## Example Usage

