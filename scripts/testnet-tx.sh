#!/usr/bin/env sh
osmosisd tx gamm \
  swap-exact-amount-in 140530uosmo 1 \
  --swap-route-pool-ids 75 \
  --swap-route-denoms factory/osmo1qc5pen6am58wxuj58vw97m72vv5tp74remsul7/uosmoexp \
  --node https://rpc.osmotest5.osmosis.zone:443 \
  --chain-id osmo-test-5 \
  --gas auto \
  --gas-adjustment 1.2 \
  --keyring-backend pass \
  --gas-prices 0.0025uosmo \
  --from bot-1 \
  --yes
