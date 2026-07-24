# static-analysis

The `static-analysis` collect plugin executes static analysis tools (PHPStan, ESLint, Pylint, or custom tools) and returns structured results including issues found, execution status, and metadata.

## Plugin fields

| Field            | Type              | Required | Default | Description                                                                 |
| ---------------- | ----------------- | :------: | :-----: | --------------------------------------------------------------------------- |
| tool             | string            |   Yes    |   ""    | The static analysis tool to use: `phpstan`, `eslint`, `pylint`, or `custom` |
| binary           | string            |    No    |   ""    | Custom binary path (overrides default tool binary)                          |
| config           | string            |    No    |   ""    | Path to configuration file for the tool                                     |
| paths            | []string          |    No    |   []    | List of paths to analyze (files or directories)                             |
| args             | []string          |    No    |   []    | Additional command line arguments                                           |
| presets          | map[string]string |    No    |   {}    | Tool-specific preset configurations                                         |
| environment      | map[string]string |    No    |   {}    | Environment variables to set before execution                               |
| ignore-error     | bool              |    No    |  false  | Whether to ignore execution errors                                          |
| output-format    | string            |    No    |   ""    | Output format: `json`, `table`, `stylish`, `text` (tool-dependent)          |
| failure-patterns | []string          |    No    |   []    | Patterns in output that indicate failure                                    |

<Content :page-key="$site.pages.find(p => p.path === '/reference/common/collect.html').key"/>

## Supported Tools

### Built-in Tools

#### PHPStan
- **Binary**: `vendor/phpstan/phpstan/phpstan`
- **Default args**: `analyse --no-progress`
- **Config flag**: `--configuration=`
- **Output formats**: `json`, `table`
- **Success exit codes**: `0` (exit code `1` analyzed for issues vs errors)

#### ESLint
- **Binary**: `npx eslint`
- **Config flag**: `--config `
- **Output formats**: `json`, `stylish`
- **Success exit codes**: `0` (exit code `1` analyzed for issues vs errors)

#### Pylint
- **Binary**: `pylint`
- **Config flag**: `--rcfile=`
- **Output formats**: `json`, `text`
- **Success exit codes**: `0`, `4`, `8`, `16`

### Custom Tools
Set `tool: custom` and provide a `binary` to use any static analysis tool.

## Return format

A JSON object with the following structure:

```json
{
  "success": true,
  "exit_code": 0,
  "output": "tool output",
  "error_output": "tool stderr",
  "issues": [
    {
      "file": "path/to/file.php",
      "line": 42,
      "column": 10,
      "message": "Issue description",
      "rule": "rule-name",
      "severity": "error"
    }
  ],
  "tool": "phpstan",
  "duration": "2.5s"
}
```

| Field        | Type    | Description                                 |
| ------------ | ------- | ------------------------------------------- |
| success      | bool    | Whether the analysis completed successfully |
| exit_code    | int     | Tool exit code                              |
| output       | string  | Standard output from the tool               |
| error_output | string  | Standard error from the tool                |
| issues       | []Issue | Parsed issues (when output format is JSON)  |
| tool         | string  | Name of the tool used                       |
| duration     | string  | Execution duration                          |

### Issue Structure

| Field    | Type   | Description                     |
| -------- | ------ | ------------------------------- |
| file     | string | Path to the file with the issue |
| line     | int    | Line number                     |
| column   | int    | Column number (optional)        |
| message  | string | Issue description               |
| rule     | string | Rule/checker name (optional)    |
| severity | string | Issue severity (optional)       |

## Examples

### Basic PHPStan Analysis

```yaml
collect:
  phpstan-results:
    static-analysis:
      tool: phpstan
      config: phpstan.neon
      paths:
        - src/
        - web/modules/custom/
      output-format: json
```

### ESLint with Custom Configuration

```yaml
collect:
  eslint-results:
    static-analysis:
      tool: eslint
      config: .eslintrc.json
      paths:
        - "src/**/*.js"
        - "src/**/*.ts"
      output-format: json
      environment:
        NODE_ENV: production
```

### Custom Tool Usage

```yaml
collect:
  custom-analysis:
    static-analysis:
      tool: custom
      binary: ./bin/my-analyzer
      args:
        - "--strict"
        - "--format=json"
      paths:
        - src/
      ignore-error: true
```

### Advanced PHPStan Configuration

```yaml
collect:
  phpstan-advanced:
    static-analysis:
      tool: phpstan
      config: "${PHPSTAN_CONFIG:-phpstan.neon}"
      paths:
        - src/
      presets:
        memory-limit: "1G"
        level: "8"
      environment:
        XDEBUG_MODE: "off"
      failure-patterns:
        - "Parse error"
        - "Fatal error"
```

## Environment Resolution

The plugin supports environment variable resolution in:
- `config`: Configuration file paths
- `paths`: Analysis target paths
- `presets`: Tool-specific settings
- `environment`: Environment variable values

Use standard `${VAR}` or `${VAR:-default}` syntax for environment variable substitution.

## Exit Code Handling

The plugin intelligently handles exit codes:

- **Exit code 0**: Always success
- **Exit code 1**: Analyzed to distinguish between "issues found" vs "execution error"
- **Other codes**: Tool-specific handling (e.g., Pylint's multiple success codes)

For exit code 1, the plugin examines output content to determine if it represents successful analysis with issues or an actual execution failure.
