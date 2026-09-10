package main

// These are meaningful example task data, not results from a real investigation.
func pricingConstraints() any {
	return struct {
		Currency         string   `json:"currency"`
		FreeShippingFrom string   `json:"free_shipping_from"`
		Rules            []string `json:"rules"`
	}{"USD", "50.00", []string{
		"Use integer cents.",
		"Charge 4.99 below the threshold.",
		"Preserve the exact price boundary.",
	}}
}

func researchNotes() any {
	return struct {
		Summary      string   `json:"summary"`
		Observations []string `json:"observations"`
	}{
		Summary: "Example research into the quote command's pricing boundary and evidence.",
		Observations: []string{
			"The boundary is inclusive: a subtotal of 50.00 receives free shipping. The adjacent inputs 49.99 and 50.01 distinguish the declared rule from a mistaken strict-greater-than comparison.",
			"Represent amounts as integer cents so decimal parsing, comparison, shipping, and total calculation do not depend on binary floating-point rounding.",
			"Below the threshold, shipping is 4.99. A 49.99 subtotal therefore produces a 54.98 total. At 50.00 and 50.01, shipping is zero and total equals subtotal.",
			"A useful demonstration invokes the actual CLI separately for all three boundary inputs and records stdout, stderr, and the exit code for every invocation.",
			"The verifier should inspect the returned monetary fields, not infer the price rule merely from a successful process exit or the presence of a test report.",
		},
	}
}

type contextNote struct {
	Key   string
	Value string
}

// Each value fits the individual limit; together they exceed the prompt budget.
func acceptanceNotes() []contextNote {
	return []contextNote{
		{"below-boundary", "For subtotal 49.99, expect shipping 4.99 and total 54.98. Capture all three monetary fields from a real CLI invocation."},
		{"exact-boundary", "For subtotal 50.00, expect shipping 0.00 and total 50.00. This exact input distinguishes inclusive from exclusive free shipping."},
		{"above-boundary", "For subtotal 50.01, expect shipping 0.00 and total 50.01. Keep this observation beside the exact-boundary and below-boundary results."},
		{"numeric-representation", "Parse decimal amounts into integer cents. Compare subtotal against 5000 cents and add shipping without floating-point arithmetic."},
		{"output-contract", "Return a JSON object with subtotal, shipping, and total fields, each formatted with exactly two decimal places for inspection."},
		{"failure-contract", "Reject malformed or negative amounts with a nonzero exit and an explanation on stderr. Do not emit a successful quote for invalid input."},
		{"independent-review", "The tester operates the command directly and inspects output fields. The implementer's report supplies navigation, not the verdict."},
		{"evidence-record", "Keep the command, exit code, stdout, and stderr for each demonstrated case so a later reviewer can inspect what actually happened."},
		{"scope-boundary", "Demonstrate the declared pricing behavior. Tax calculation, currency conversion, discounts, and network requests are outside this task."},
	}
}
