#!/usr/bin/env sh
osmosisd tx gamm \
  swap-exact-amount-in 140530uosmo 1 \
  --swap-route-pool-ids 118 \
  --swap-route-denoms ibc/A8C2D23A1E6F95DA4E48BA349667E322BD7A6C996D8A4AAE8BA72E190F3D1477 \
  --node https://rpc.osmotest5.osmosis.zone:443 \
  --chain-id osmo-test-5 \
  --gas auto \
  --gas-adjustment 1.2 \
  --keyring-backend pass \
  --gas-prices 0.0025uosmo \
  --from bot-1 \
  --yes
