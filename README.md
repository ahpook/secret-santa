- what exactly i've done to make money present in backend wallet:
    ```
    neo-go wallet nep17 transfer --from Nhfg3TbpwogLvDGVvAvqyThbsHgoSUKwtn --to AddressOfBrandNewCreatedAccountInCreatedWallet --token GAS --amount 2000 -w /path/to/your/frostfs-aio local directory/morph/node-wallet.json -r http://localhost:30333
    ```

    - `Nhfg3TbpwogLvDGVvAvqyThbsHgoSUKwtn` - an address the most important 'morph' wallet with an 'infinite' amount of money
    - `AddressOfBrandNewCreatedAccountInCreatedWallet` - just an address in `wallet.json`

- how i deployed a contract:
    - i've deployed a 'goofyahhdocuments' contract from a frostfs-aio directory, BUT the path to the wallet was to brand new backend wallet:
        ```
        dimonlimon@MagicBook16:~/smart-contracts/frostfs-aio$ neo-go contract deploy -i goofyahhdocuments/main.nef -m goofyahhdocuments/main.manifest.json -r  http://localhost:30333 -w /home/dimonlimon/secret-santa/backend/wallet.json
        Enter account NQmsTaeU43gbowXNyP8oj2wE9dreUPCNeP password > 
        Network fee: 0.0177052
        System fee: 10.0106065
        Total fee: 10.0283117
        Relay transaction (y|N)> y
        6083fac32c93cf298cda701971ac33ef8b0a5b1be857073649f8c77088460d1e
        Contract: d487b1f344c783d9275fbc3fcc15fc5209aa6388 // <- this contract hash goes directly to the config.yaml's 'goofyahhdocuments_contract' field
        ```
    

---

- commands to interact with contracts, wallets
	- `make up tick.epoch`
	- `neo-go contract init --name test`
	- `neo-go contract compile --in goofyahhcontract/main.go`
	- check wallet: `neo-go wallet nep17 balance -r http://localhost:30333 -w ~/frostfs-aio/morph/node-wallet.json`
	- `neo-go wallet nep17 transfer --from ... --to ... --token GAS --amount 20 -w morph/node-wallet.json -r http://localhost:30333`
	- `neo-go contract deploy -i secret_santa/secret_santa.nef -m secret_santa/secret_santa.manifest.json -r  http://localhost:30333 -w wallets/wallet1.json`
		- `curl -s --data '{"id":1,"jsonrpc":"2.0","method":"getapplicationlog","params":["4f31ef755abe071114b12b8d2f021308c5ed5f55bdbe9992d89830f6b7111203"]}' http://localhost:30333 | jq`