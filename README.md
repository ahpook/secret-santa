- что я сделал, чтобы на новосозданном серверном кошельке были деньги:
    ```
    neo-go wallet nep17 transfer --from Nhfg3TbpwogLvDGVvAvqyThbsHgoSUKwtn --to AddressOfBrandNewCreatedAccountInCreatedWallet --token GAS --amount 2000 -w /path/to/your/frostfs-aio local directory/morph/node-wallet.json -r http://localhost:30333
    ```

    - `Nhfg3TbpwogLvDGVvAvqyThbsHgoSUKwtn` - адрес самого главного morph кошелька с бесконечным колвом денег
    - `AddressOfBrandNewCreatedAccountInCreatedWallet` - просто адрес аккаунта в `wallet.json`