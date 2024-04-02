#!/usr/bin/env sh
# Grant bot-1 authz from bot-5

# Define grantee address and other constants
GRANTEE="osmo19z9j3240alpa2xl0h4epvr5vryjmam6w3pwm4h"
# NODE="https://rpc.osmotest5.osmosis.zone:443"
NODE="https://osmosis-testnet-rpc.polkachu.com:443"
CHAIN_ID="osmo-test-5"
GAS="auto"
GAS_ADJUSTMENT="1.2"
KEYRING_BACKEND="pass"
GAS_PRICES="0.0025uosmo"
FROM="bot-5"
EXPIRATION=$(date -d "+1 year" +%s)

# Message types separated by spaces (so it is POSIX compliant)
MSG_TYPES="/osmosis.concentratedliquidity.v1beta1.MsgCreatePosition /osmosis.concentratedliquidity.v1beta1.MsgWithdrawPosition"

# Loop through the message types
for MSG_TYPE in $MSG_TYPES; do
    osmosisd tx authz grant \
      $GRANTEE \
      generic \
      --msg-type="$MSG_TYPE" \
      --node $NODE \
      --chain-id $CHAIN_ID \
      --gas $GAS \
      --gas-adjustment $GAS_ADJUSTMENT \
      --keyring-backend $KEYRING_BACKEND \
      --gas-prices $GAS_PRICES \
      --from $FROM \
      --expiration $EXPIRATION \
      --yes
    sleep 5
done
