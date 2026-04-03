package model

// ModelPricing define preços por milhão de tokens para cada modelo.
type ModelPricing struct {
	InputPerMillion       float64
	OutputPerMillion      float64
	CacheReadPerMillion   float64
	CacheWritePerMillion  float64
}

// Tabela de preços atualizada (abril 2026).
var PricingTable = map[string]ModelPricing{
	"claude-opus-4-6": {
		InputPerMillion:      15.0,
		OutputPerMillion:     75.0,
		CacheReadPerMillion:  1.5,
		CacheWritePerMillion: 18.75,
	},
	"claude-opus-4-5-20251101": {
		InputPerMillion:      15.0,
		OutputPerMillion:     75.0,
		CacheReadPerMillion:  1.5,
		CacheWritePerMillion: 18.75,
	},
	"claude-sonnet-4-6": {
		InputPerMillion:      3.0,
		OutputPerMillion:     15.0,
		CacheReadPerMillion:  0.3,
		CacheWritePerMillion: 3.75,
	},
	"claude-sonnet-4-5-20251022": {
		InputPerMillion:      3.0,
		OutputPerMillion:     15.0,
		CacheReadPerMillion:  0.3,
		CacheWritePerMillion: 3.75,
	},
	"claude-haiku-4-5-20251001": {
		InputPerMillion:      0.80,
		OutputPerMillion:     4.0,
		CacheReadPerMillion:  0.08,
		CacheWritePerMillion: 1.0,
	},
	"claude-3-5-sonnet-20241022": {
		InputPerMillion:      3.0,
		OutputPerMillion:     15.0,
		CacheReadPerMillion:  0.3,
		CacheWritePerMillion: 3.75,
	},
	"claude-3-5-haiku-20241022": {
		InputPerMillion:      0.80,
		OutputPerMillion:     4.0,
		CacheReadPerMillion:  0.08,
		CacheWritePerMillion: 1.0,
	},
}

// GetPricing retorna o pricing do modelo, com fallback por prefixo.
func GetPricing(model string) ModelPricing {
	if p, ok := PricingTable[model]; ok {
		return p
	}
	// Fallback por prefixo
	prefixes := []struct {
		prefix  string
		pricing ModelPricing
	}{
		{"claude-opus", PricingTable["claude-opus-4-6"]},
		{"claude-sonnet", PricingTable["claude-sonnet-4-6"]},
		{"claude-haiku", PricingTable["claude-haiku-4-5-20251001"]},
		{"claude-3-5-sonnet", PricingTable["claude-3-5-sonnet-20241022"]},
		{"claude-3-5-haiku", PricingTable["claude-3-5-haiku-20241022"]},
	}
	for _, p := range prefixes {
		if len(model) >= len(p.prefix) && model[:len(p.prefix)] == p.prefix {
			return p.pricing
		}
	}
	// Fallback final: Sonnet (preço médio)
	return PricingTable["claude-sonnet-4-6"]
}

// CalculateCost calcula o custo total de um TokenUsage com pricing do modelo.
func CalculateCost(model string, tokens TokenUsage) float64 {
	p := GetPricing(model)
	cost := float64(tokens.InputTokens) * p.InputPerMillion / 1_000_000
	cost += float64(tokens.OutputTokens) * p.OutputPerMillion / 1_000_000
	cost += float64(tokens.CacheReadTokens) * p.CacheReadPerMillion / 1_000_000
	cost += float64(tokens.CacheWriteTokens) * p.CacheWritePerMillion / 1_000_000
	return cost
}

// CalculateSessionCost calcula o custo total de uma sessão usando preço correto por modelo.
func CalculateSessionCost(modelTokens map[string]TokenUsage) float64 {
	var total float64
	for model, tokens := range modelTokens {
		total += CalculateCost(model, tokens)
	}
	return total
}
