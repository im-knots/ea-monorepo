# Ea Credentials Manager

## API Documentation
### Required Headers
All requests to this API coming into the cluster via the api gateway must include an authorization header containing an authenticated user's JWT

```
Authorization: Bearer <YOUR JWT>
```

Internal systems within the cluster (behind Kong) can access this service by providing

(**Note: network level access is restricted in the cluster via NetworkPolicies**)

```
x-consumer-username: internal
```


---