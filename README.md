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
    Contract: d487b1f344c783d9275fbc3fcc15fc5209aa6388
    ```