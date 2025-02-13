# command

The `command` collect plugin executes a command and returns the output, error and exit code as a map.

## Plugin fields

| Field        | Description                                           | Required | Default |
| ------------ | ----------------------------------------------------- | :------: | :-----: |
| cmd          | The main command to run.                              |   Yes    |   ""    |
| args         | A list of arguments to pass to the command.           |    No    |   []    |
| ignore-error | Whether to fail data collection if the command fails. |    No    |  false  |

<Content :page-key="$site.pages.find(p => p.path === '/reference/common/collect.html').key"/>

## Return format

A map with the following fields:

| Field | Description            |
| ----- | ---------------------- |
| out   | The command output.    |
| err   | The command error.     |
| code  | The command exit code. |

## Example

<<< @/../examples/remediation.yml{3-9}
