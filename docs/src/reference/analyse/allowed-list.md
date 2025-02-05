# allowed:list

The `allowed:list` plugin checks if values are allowed for a number of input formats.

## Plugin fields

::: warning
TODO: Add details on how the logic works.
:::

| Field         | Description                    | Required | Default |
| ------------- | ------------------------------ | :------: | :-----: |
| package-match | The package to match against.  |    No    |   ""    |
| allowed       | The list of allowed values.    |    No    |   []    |
| required      | The list of required values.   |    No    |   []    |
| deprecated    | The list of deprecated values. |    No    |   []    |
| exclude-keys  | The list of keys to exclude.   |    No    |   []    |
| ignore        | The list of values to ignore.  |    No    |   []    |


<Content :page-key="$site.pages.find(p => p.path === '/reference/common/analyse.html').key"/>
