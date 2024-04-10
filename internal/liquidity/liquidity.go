package liquidity

import (
	"errors"
	"fmt"

	"github.com/osmosis-labs/osmosis/osmomath"
	clmath "github.com/osmosis-labs/osmosis/v24/x/concentrated-liquidity/math"
	"github.com/osmosis-labs/osmosis/v24/x/concentrated-liquidity/model"
	cltypes "github.com/osmosis-labs/osmosis/v24/x/concentrated-liquidity/types"
	"go.uber.org/zap"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TickSpacing = int64(100)

// createPositionMsg creates a new CL position message
func createPositionMsg(poolID uint64, lowerTick, upperTick int64, tokens sdk.Coins, addr string, _ bool) sdk.Msg {
	var amount0, amount1 sdkmath.Int
	amount0, amount1 = sdk.ZeroInt(), sdk.ZeroInt()

	// Generate the swap message
	msg := cltypes.MsgCreatePosition{
		PoolId:          poolID,
		Sender:          addr,
		LowerTick:       lowerTick,
		UpperTick:       upperTick,
		TokensProvided:  tokens,
		TokenMinAmount0: amount0,
		TokenMinAmount1: amount1,
	}

	return &msg
}

// removePositionMsg withdraws positions with specific ids
func removePositionMsg(position model.Position) sdk.Msg {
	// Generate the swap message
	msg := cltypes.MsgWithdrawPosition{
		PositionId:      position.PositionId,
		Sender:          position.Address,
		LiquidityAmount: position.Liquidity,
	}

	return &msg
}

// marketMake creates a market making positions
func RemovePreviousPositions(l *zap.Logger, positions []model.FullPositionBreakdown) []sdk.Msg {
	var msgs []sdk.Msg

	for _, p := range positions {
		l.Debug("position",
			zap.Uint64("positionId", p.Position.PositionId),
			zap.String("liquidity", p.Position.Liquidity.String()),
		)

		msg := removePositionMsg(p.Position)

		msgs = append(msgs, msg)
	}

	return msgs
}

// marketMake creates a market making positions
func MarketMake(l *zap.Logger, poolID uint64, currentTick int64, spotPrice, targetPrice, spread string, token0 sdk.Coin, token1 sdk.Coin, addr string) ([]sdk.Msg, error) {
	l.Debug("inputs",
		zap.String("spotPrice", spotPrice),
		zap.String("targetPrice", targetPrice),
	)

	spotPriceAsBigDec, err := osmomath.NewBigDecFromStr(spotPrice)
	if err != nil {
		l.Error("Failed to convert spot price to big dec", zap.Error(err))
		return nil, err
	}

	targetPriceAsBigDec, err := osmomath.NewBigDecFromStr(targetPrice)
	if err != nil {
		l.Error("Failed to convert target price to big dec", zap.Error(err))
		return nil, err
	}

	spreadAsBigDec, err := osmomath.NewBigDecFromStr(spread)
	if err != nil {
		l.Error("Failed to convert spread to big dec", zap.Error(err))
		return nil, err
	}

	if spotPriceAsBigDec.LT(targetPriceAsBigDec) {
		targetPriceAsBigDec, spotPriceAsBigDec = spotPriceAsBigDec, targetPriceAsBigDec
	}

	buyTick, lowTick, sellTick, highTick, err := calculateBuySellTicks(l, targetPriceAsBigDec, spotPriceAsBigDec, spreadAsBigDec)
	if err != nil {
		l.Error("Failed to calculate buy and sell ticks", zap.Error(err))
		return nil, err
	}

	lowTick, buyTick = adjustForCurrentTick(l, true, currentTick, lowTick, buyTick)
	sellTick, highTick = adjustForCurrentTick(l, false, currentTick, sellTick, highTick)

	if !(lowTick < buyTick && buyTick < sellTick && sellTick < highTick) {
		err := errors.New("ticks are in the incorrect order")
		l.Error("Failed to calculate buy and sell ticks", zap.Error(err))
		return nil, err
	}

	buyPosition := createPositionMsg(poolID, lowTick, buyTick, sdk.NewCoins(token1), addr, true)
	sellPosition := createPositionMsg(poolID, sellTick, highTick, sdk.NewCoins(token0), addr, false)

	fmt.Println("buyPosition", buyPosition)
	fmt.Println("sellPosition", sellPosition)

	return []sdk.Msg{buyPosition, sellPosition}, nil
}

func adjustForCurrentTick(l *zap.Logger, isBuy bool, currentTick, lowerTick, upperTick int64) (int64, int64) {
	if lowerTick <= currentTick && currentTick <= upperTick {
		lowerTick, upperTick = adjustTicksWithinRange(isBuy, currentTick, lowerTick, upperTick)
	}

	upperTick = roundDownTick(l, upperTick)
	lowerTick = roundDownTick(l, lowerTick)

	lowerTick = adjustLowerTick(currentTick, lowerTick)

	return lowerTick, upperTick
}

func adjustTicksWithinRange(isBuy bool, currentTick, lowerTick, upperTick int64) (int64, int64) {
	if isBuy {
		upperTick = currentTick - TickSpacing
	} else {
		lowerTick = currentTick + TickSpacing
	}
	return lowerTick, upperTick
}

func roundDownTick(l *zap.Logger, tick int64) int64 {
	roundedTick, err := clmath.RoundDownTickToSpacing(tick, TickSpacing)
	if err != nil {
		l.Error("Failed to round down tick", zap.Error(err))
	}
	return roundedTick
}

func adjustLowerTick(currentTick, lowerTick int64) int64 {
	lowerDelta := abs(currentTick - lowerTick)
	if lowerDelta < TickSpacing {
		lowerTick += (3 * TickSpacing)
	}
	return lowerTick
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// calculateBuySellTicks calculates the buy and sell ticks
// TODO: we should return errors...
func calculateBuySellTicks(l *zap.Logger, buyPrice, sellPrice, spread osmomath.BigDec) (int64, int64, int64, int64, error) { //nolint:unparam
	// get the lower and upper bounds
	buyLowerBound := buyPrice.Mul(osmomath.OneBigDec().Sub(spread))
	sellUpperBound := sellPrice.Mul(osmomath.OneBigDec().Add(spread))

	// Calculate the buy and sell ticks
	buyPriceTick, err := calculateAndRoundPriceToTick(buyPrice)
	if err != nil {
		l.Error("Failed to calculate buy price tick", zap.Error(err))
	}

	buyLowerTick, err := calculateAndRoundPriceToTick(buyLowerBound)
	if err != nil {
		l.Error("Failed to calculate buy lower bound price tick", zap.Error(err))
	}

	sellPriceTick, err := calculateAndRoundPriceToTick(sellPrice)
	if err != nil {
		l.Error("Failed to calculate sell price tick", zap.Error(err))
	}

	sellUpperTick, err := calculateAndRoundPriceToTick(sellUpperBound)
	if err != nil {
		l.Error("Failed to calculate sell upper bound price tick", zap.Error(err))
	}

	return buyPriceTick, buyLowerTick, sellPriceTick, sellUpperTick, nil
}

func calculateAndRoundPriceToTick(price osmomath.BigDec) (int64, error) {
	priceTick, err := clmath.CalculatePriceToTick(price)
	if err != nil {
		return 0, err
	}

	priceTick, err = clmath.RoundDownTickToSpacing(priceTick, TickSpacing)
	if err != nil {
		return 0, err
	}

	return priceTick, nil
}
