## Common fields

| Field             | Description                                                                                         | Required | Default |
| ----------------- | --------------------------------------------------------------------------------------------------- | :------: | :-----: |
| name              | The name/identifier of the plugin - this is the yaml key in the config file when defining the fact. |   Yes    |    -    |
| connection        | The connection to use for collecting the fact.                                                      |    No    |   ""    |
| input             | A previous input to use when collecting the fact.                                                   |    No    |   ""    |
| additional-inputs | Additional previous inputs to use when collecting the fact.                                         |    No    |   []    |
