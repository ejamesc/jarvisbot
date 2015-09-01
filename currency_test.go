package jarvisbot_test

import (
	"testing"

	"github.com/ejamesc/jarvisbot"
)

func TestParseArgs(t *testing.T) {
	res, err := jarvisbot.RetrieveExchangeRates()

	ok(t, err)
	assert(t, res.UnixTimestamp != 0, "Exchange rate timestamp should not be empty.")
	assert(t, res.Base != "", "Exchange rate base should not be empty.")
	assert(t, res.Rates != nil, "Exchange rates should not be empty.")
}
