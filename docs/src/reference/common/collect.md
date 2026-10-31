## Common fields

| Field             | Description                                                                                         | Required | Default |
| ----------------- | --------------------------------------------------------------------------------------------------- | :------: | :-----: |
| name              | The name/identifier of the plugin - this is the yaml key in the config file when defining the fact. |   Yes    |    -    |
| connection        | The connection to use for collecting the fact.                                                      |    No    |   ""    |
| input             | A previous input to use when collecting the fact.                                                   |    No    |   ""    |
| additional-inputs | Additional previous inputs to use when collecting the fact.                                         |    No    |   []    |
| category          | The Lagoon fact category, used only by `shipshape collect -o lagoon-facts`.                         |    No    | fact id |
| description       | The Lagoon fact description, used only by `shipshape collect -o lagoon-facts`.                      |    No    |   ""    |

### Lagoon facts output

`shipshape collect -o lagoon-facts` emits collected facts as JSON matching the
Lagoon `addFactsByName` mutation's `input` object, instead of the default
human-readable `pretty` output:

```sh
LAGOON_PROJECT=my-project LAGOON_ENVIRONMENT=main \
  shipshape collect -f shipshape.yml -o lagoon-facts | jq .
```

This only emits the payload - **no API call is made**, so no Lagoon API
credentials are required or read. `project`/`environment` come from
`LAGOON_PROJECT`/`LAGOON_ENVIRONMENT` (the `--lagoon-project`/
`--lagoon-environment` flags are not registered on `collect`, only on `run`),
and the command fails fast if either is unset. `source` comes from the
`output.lagoon.source` config value (no env var equivalent; `--lagoon-source`
is likewise not registered on `collect`), falling back to `"Shipshape"`.

Pipe the output into a GraphQL client to actually push facts to Lagoon, e.g.:

```sh
shipshape collect -f shipshape.yml -o lagoon-facts > facts.json
curl -s "$LAGOON_API" \
  -H "Authorization: Bearer $LAGOON_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"query\": \"mutation(\$input: AddFactsByNameInput!) { addFactsByName(input: \$input) { id } }\", \"variables\": {\"input\": $(cat facts.json)}}"
```

The payload may carry collected fact data (e.g. database contents from
`database:search`) - treat `-o lagoon-facts` output at the classification of
the data collected, and avoid echoing it into public CI logs.
