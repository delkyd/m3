	"github.com/m3db/m3x/instrument"
	"github.com/m3db/m3db/ratelimit"
	// SetRateLimitOptions sets the rate limit options
	SetRateLimitOptions(value ratelimit.Options) Options

	// RateLimitOptions returns the rate limit options
	RateLimitOptions() ratelimit.Options
