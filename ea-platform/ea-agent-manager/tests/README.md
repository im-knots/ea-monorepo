# Ea Agent Manager Tests
Here you will find some tests for the Ea Agent Manager. 

## Create a new API smoke test
- Add the json payload you are sending for PUT operations into the `smoke/payloads` directory. Name it `test-<APIhandler>-#.json`
- Add a new smoke test to this `smoke` directory. Use existing examples to create a bash script to iterate over your test payloads and curl them to the endpoint you are testing. 


## Run API smoke tests

```bash
./smoke/create-agent.sh
./smoke/create-node.sh
./smoke/get-all-agents.sh
./smoke/get-all-nodes.sh
./smoke/get-agent.sh
./smoke/get-node.sh
```
