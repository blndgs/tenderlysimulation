# 4337 userop simulation

## Set up env values

```sh
make env
```

- `ETH_RPC_URL`: the url for your rpc endpoint
- `CHAIN_ID`: the chain id for the chain. Defaults to 888

## Run

```~~
make run
```

This would generate a new throwaway Ethereum wallet, create a very basic userop,
sign it and try to run a simulation against the RPC
