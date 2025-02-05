## Common fields

| Field         | Description                                                                                                | Required |        Default        |
| ------------- | ---------------------------------------------------------------------------------------------------------- | :------: | :-------------------: |
| name          | The name of the policy - this is the yaml key in the config file when defining the policy.                 |   Yes    |           -           |
| description   | The description of the policy - if specified, it will be used as the heading for the policy in the output. |    No    |          ""           |
| input         | The input for the policy - used to select the fact plugin to use.                                          |   Yes    |           -           |
| severity      | The severity of the policy when breached (low, normal, high, critical)                                     |    No    |        normal         |
| breach-format | The breach template for the policy. The table below shows the available fields.                            |    No    | Empty breach template |

### Breach template

::: warning
TODO: Add information on how to use go template variables.
:::

| Field       | Description              | Required | Default |
| ----------- | ------------------------ | :------: | :-----: |
| type        | The type of breach.      |   Yes    |   ""    |
| key-label   | The label for the key.   |    No    |   ""    |
| key         | The key.                 |    No    |   ""    |
| value-label | The label for the value. |    No    |   ""    |
| value       | The value.               |    No    |   ""    |
