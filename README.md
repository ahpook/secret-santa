# Steps to run this project
## From frostfs-aio local repo:
- `make image-aio`
- `make up`
- `docker cp user-cli-cfg.yaml aio:/config/user-cli-cfg.yaml`
- `docker exec aio frostfs-cli container create -c /config/user-cli-cfg.yaml --policy 'REP 1' --await`
- `docker exec aio frostfs-cli -c /config/user-cli-cfg.yaml ape-manager add --chain-id nyan --rule 'Allow Object.* *' --target-type container --target-name ...`

## From our github repo root directory:
- `neo-go contract compile --in documentsdistributor/main.go`
- if there are no money on the wallet -> transfer GAS to it `neo-go wallet nep17 transfer --from ... --to ... --token GAS --amount 20 -w morph/node-wallet.json -r http://localhost:30333`
- `neo-go contract deploy -i documentsdistributor/main.nef -m documentsdistributor/main.manifest.json -r  http://localhost:30333 -w backend/wallet.json`

## Additional commands to debug:
- `curl -s --data '{"id":1,"jsonrpc":"2.0","method":"getapplicationlog","params":[""]}' http://localhost:30333 | jq`
- `frostfs-cli -c /config/user-cli-cfg.yaml object put --cid HRnXPvSEuaZttVSmSADKJvvbBsp9o4swKEtmnD2avVxp --file goofyahhdocument.txt`

## How to run backend code:
- `cd backend`
- `go run main.go config.yaml`

## How to run client code:
- `cd client`
- `go run main.go`
