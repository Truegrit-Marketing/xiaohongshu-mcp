package humanize

import (
	"context"
	"time"
)

// paceScale 停顿的倍数：1 是照常；小于 1 是提速（XHS_PACE_SCALE，见 configs.PaceScaleFromEnv）。
// 只缩「像人一样停一下」的停顿，鼠标轨迹、打字节奏不动。
var paceScale = 1.0

// SetPaceScale 设停顿的倍数；不在 (0, 1] 里的不收。
func SetPaceScale(scale float64) {
	if scale > 0 && scale <= 1 {
		paceScale = scale
	}
}

// Scaled 按停顿的倍数缩短一段时间（给 humanize 外面的固定停顿用）。
func Scaled(d time.Duration) time.Duration {
	return time.Duration(float64(d) * paceScale)
}

func Delay(ctx context.Context, action Action) {
	dist, ok := defaultProvider.Timing()[action]
	if !ok {
		dist = defaultProvider.Timing()[AfterClick]
	}

	t := time.NewTimer(Scaled(dist.Sample()))
	defer t.Stop()

	select {
	case <-t.C:
	case <-ctx.Done():
	}
}
