package traffic

import (
	"fmt"
	"github.com/Cleamy/uy_micro/config"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/core/system"
)

func InitSentinel(cfg *config.SentinelConfig) error {
	if !cfg.Enable {
		return nil
	}
	if err := sentinel.InitDefault(); err != nil {
		return fmt.Errorf("sentinel core init err: %w", err)
	}

	// 1. 加载流控规则 flow.Rule
	if len(cfg.FlowRules) > 0 {
		var rules []*flow.Rule
		for _, r := range cfg.FlowRules {
			rules = append(rules, &flow.Rule{
				ID:                     r.ID,
				Resource:               r.Resource,
				Threshold:              r.Threshold,
				StatIntervalInMs:       r.StatIntervalInMs,
				ControlBehavior:        flow.ControlBehavior(r.ControlBehavior),
				MaxQueueingTimeMs:      r.MaxQueueingTimeMs,
				TokenCalculateStrategy: flow.TokenCalculateStrategy(r.TokenCalculateStrategy),
				RelationStrategy:       flow.RelationStrategy(r.RelationStrategy),
				RefResource:            r.RefResource,
				WarmUpPeriodSec:        r.WarmUpPeriodSec,
				WarmUpColdFactor:       r.WarmUpColdFactor,
				LowMemUsageThreshold:   r.LowMemUsageThreshold,
				HighMemUsageThreshold:  r.HighMemUsageThreshold,
				MemLowWaterMarkBytes:   r.MemLowWaterMarkBytes,
				MemHighWaterMarkBytes:  r.MemHighWaterMarkBytes,
			})
		}
		if _, err := flow.LoadRules(rules); err != nil {
			return fmt.Errorf("load flow rule err: %w", err)
		}
	}

	// 2. 加载熔断降级规则 circuitbreaker.Rule
	if len(cfg.CircuitBreakerRules) > 0 {
		var rules []*circuitbreaker.Rule
		for _, r := range cfg.CircuitBreakerRules {
			rules = append(rules, &circuitbreaker.Rule{
				Id:               r.ID,
				Resource:         r.Resource,
				Strategy:         circuitbreaker.Strategy(r.Strategy),
				Threshold:        r.Threshold,
				RetryTimeoutMs:   r.RetryTimeoutMs,
				MinRequestAmount: r.MinRequestAmount,
				StatIntervalMs:   r.StatIntervalMs,
				MaxAllowedRtMs:   r.MaxAllowedRtMs,
			})
		}
		if _, err := circuitbreaker.LoadRules(rules); err != nil {
			return fmt.Errorf("load circuit breaker rule err: %w", err)
		}
	}

	// 3. 加载热点参数限流规则 hotspot.Rule
	if len(cfg.HotSpotRules) > 0 {
		var rules []*hotspot.Rule
		for _, r := range cfg.HotSpotRules {
			rules = append(rules, &hotspot.Rule{
				ID:                r.ID,
				Resource:          r.Resource,
				MetricType:        hotspot.MetricType(r.MetricType),
				ControlBehavior:   hotspot.ControlBehavior(r.ControlBehavior),
				ParamIndex:        r.ParamIndex,
				ParamKey:          r.ParamKey,
				Threshold:         r.Threshold,
				MaxQueueingTimeMs: r.MaxQueueingTimeMs,
				BurstCount:        r.BurstCount,
				DurationInSec:     r.DurationInSec,
				ParamsMaxCapacity: r.ParamsMaxCapacity,
				SpecificItems:     r.SpecificItems,
			})
		}
		if _, err := hotspot.LoadRules(rules); err != nil {
			return fmt.Errorf("load hotspot rule err: %w", err)
		}
	}

	// 4. 加载系统保护规则
	if len(cfg.SystemRules) > 0 {
		var rules []*system.Rule
		for _, r := range cfg.SystemRules {
			rules = append(rules, &system.Rule{
				TriggerCount: r.TriggerCount,
			})
		}
		if _, err := system.LoadRules(rules); err != nil {
			return fmt.Errorf("load system rule err: %w", err)
		}
	}

	return nil
}
