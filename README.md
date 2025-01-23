### Commands to interact with contracts, wallets
	- `make up tick.epoch`
	- `neo-go contract init --name test`
	- `neo-go contract compile --in documentsdistributor/main.go`
	- check wallet: `neo-go wallet nep17 balance -r http://localhost:30333 -w ~/frostfs-aio/morph/node-wallet.json`
	- `neo-go wallet nep17 transfer --from ... --to ... --token GAS --amount 20 -w morph/node-wallet.json -r http://localhost:30333`
	- `neo-go contract deploy -i documentsdistributor/main.nef -m documentsdistributor/main.manifest.json -r  http://localhost:30333 -w wallets/wallet.json`
		- `curl -s --data '{"id":1,"jsonrpc":"2.0","method":"getapplicationlog","params":[""]}' http://localhost:30333 | jq`



- `frostfs-cli -c /config/user-cli-cfg.yaml object put --cid HRnXPvSEuaZttVSmSADKJvvbBsp9o4swKEtmnD2avVxp --file goofyahhdocument.txt`