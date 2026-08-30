package config

// PromptShieldCatalog lists models from https://huggingface.co/collections/robbypambudi/prompt-shield
var PromptShieldCatalog = []PromptShieldModelInfo{
	{
		ID:          PromptShieldSmall,
		Name:        "Prompt-Shield Flan-T5 Small",
		Params:      "60.8M",
		Description: "Lightweight prompt-injection classifier for low-latency runtime defense.",
	},
	{
		ID:          PromptShieldBase,
		Name:        "Prompt-Shield Flan-T5 Base",
		Params:      "0.2B",
		Description: "Higher-capacity classifier for improved detection accuracy.",
	},
}

type PromptShieldModelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Params      string `json:"params"`
	Description string `json:"description"`
}

func IsPromptShieldModel(model string) bool {
	for _, m := range PromptShieldCatalog {
		if m.ID == model {
			return true
		}
	}
	return false
}
