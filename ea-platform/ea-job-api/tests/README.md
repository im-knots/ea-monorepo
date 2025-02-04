# Ea Job API Tests
Here you will find some tests for the Ea Job Engine API. 

## Create a new API smoke test
- Add the json payload you are sending for PUT operations into the `smoke/payloads` directory.
- Add a new smoke test to this `smoke` directory. Use existing examples to create a bash script to iterate over your test payloads and curl them to the endpoint you are testing. 


## Run API smoke tests

```bash
./smoke/create-job.sh
```

To generate n number of jobs with the same AgentID run the above commands with a for loop in the shell. ex 100
```bash
for i in {1..100}; do ./smoke/post-job.sh; done
```