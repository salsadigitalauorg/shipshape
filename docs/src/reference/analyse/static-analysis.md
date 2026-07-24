# static-analysis:breaches

The `static-analysis:breaches` analyser examines static analysis results from the `static-analysis` fact plugin and reports breaches based on execution success, issue count, severity levels, and specific rules.

## Configuration

| Field         | Type     | Required | Default | Description                                                  |
| ------------- | -------- | :------: | :-----: | ------------------------------------------------------------ |
| check-success | bool     |    No    |  true   | Whether to check if the static analysis execution succeeded  |
| check-issues  | bool     |    No    |  true   | Whether to analyze and report on issues found                |
| min-severity  | string   |    No    |   ""    | Minimum severity level to report: `error`, `warning`, `info` |
| ignore-rules  | []string |    No    |   []    | List of rule names to ignore                                 |
| max-issues    | int      |    No    |    0    | Maximum allowed issues (0 = fail on any issues)              |

<Content :page-key="$site.pages.find(p => p.path === '/reference/common/analyse.html').key"/>

## Supported Input Formats

- `FormatRaw`: JSON data from the `static-analysis` fact plugin
- `FormatString`: JSON string representation of static analysis results

## Breach Types

### Execution Failure Breaches
When `check-success` is enabled (default), reports breaches if:
- The static analysis tool failed to execute successfully
- Exit codes indicate execution errors (not just issues found)

### Issue Count Breaches
When `check-issues` is enabled (default), reports breaches based on:
- **max-issues = 0**: Any issues found (structured or unstructured output)
- **max-issues > 0**: Only when issue count exceeds the threshold

### Severity Filtering
When `min-severity` is set, only issues at or above the specified level are considered:
- `error` (highest): Only error-level issues
- `warning`: Warning and error-level issues
- `info` (lowest): All issues

### Rule Filtering
When `ignore-rules` is configured, issues from specified rules are excluded from analysis.

## Severity Levels

The analyser recognizes three severity levels (highest to lowest):

1. **error** (level 3)
2. **warning** (level 2)
3. **info** (level 1)

Unknown severity levels are treated as includable by default.

## Issue Reporting

### Structured Issues
When the static analysis tool provides structured JSON output, issues are grouped by file and reported with detailed location information:

```
PHPStan issues in src/MyClass.php:
- line 42, column 10: Variable $undefined might not be defined (phpstan.rules.variable)
- line 55: Method signature mismatch (phpstan.rules.method)
```

### Unstructured Output
When structured parsing fails but output exists, the raw tool output is reported:

```
ESLint found issues:
/src/app.js
  1:1  error  'console' is not defined  no-undef
  2:5  warning  Missing semicolon  semi
```

## Examples

### Basic Configuration

```yaml
analyse:
  phpstan-check:
    static-analysis:breaches:
      input: phpstan-results
      # Uses defaults: check-success=true, check-issues=true, max-issues=0
```

### Error-Only Analysis

```yaml
analyse:
  eslint-errors:
    static-analysis:breaches:
      input: eslint-results
      min-severity: error
      max-issues: 0
```

### Threshold-Based Analysis

```yaml
analyse:
  pylint-threshold:
    static-analysis:breaches:
      input: pylint-results
      max-issues: 10  # Only fail if more than 10 issues
      min-severity: warning
```

### Rule Filtering

```yaml
analyse:
  phpstan-filtered:
    static-analysis:breaches:
      input: phpstan-results
      ignore-rules:
        - phpstan.rules.deadCode
        - phpstan.rules.unusedVariable
      min-severity: error
```

### Execution-Only Check

```yaml
analyse:
  tool-success:
    static-analysis:breaches:
      input: custom-tool-results
      check-success: true
      check-issues: false  # Only verify tool ran successfully
```

### Custom Severity Handling

```yaml
analyse:
  strict-analysis:
    static-analysis:breaches:
      input: eslint-results
      check-success: true
      check-issues: true
      min-severity: info      # Include all severity levels
      max-issues: 0           # Fail on any issues
      ignore-rules:
        - no-console         # Allow console statements
        - prefer-const       # Allow let declarations
```

## Input Data Structure

The analyser expects input data matching the `static-analysis` fact plugin output:

```json
{
  "success": true,
  "exit_code": 1,
  "output": "tool output text",
  "error_output": "error messages",
  "issues": [
    {
      "file": "src/example.php",
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

## Common Use Cases

### CI/CD Quality Gates
Use `max-issues` thresholds to enforce code quality standards while allowing gradual improvement:

```yaml
analyse:
  quality-gate:
    static-analysis:breaches:
      input: phpstan-results
      max-issues: 50  # Allow up to 50 issues, fail beyond that
      min-severity: warning
```

### Security-Focused Analysis
Focus only on error-level issues that might indicate security problems:

```yaml
analyse:
  security-check:
    static-analysis:breaches:
      input: eslint-security-results
      min-severity: error
      max-issues: 0
      ignore-rules:
        - no-console  # Console statements aren't security issues
```

### Development vs Production
Different thresholds for different environments:

```yaml
# Development - allow more issues
analyse:
  dev-check:
    static-analysis:breaches:
      input: analysis-results
      max-issues: 20
      min-severity: warning

# Production - strict enforcement
analyse:
  prod-check:
    static-analysis:breaches:
      input: analysis-results
      max-issues: 0
      min-severity: error
```
