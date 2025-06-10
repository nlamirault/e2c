# Using Feature Flags with OpenFeature in e2c

This guide explains how to use feature flags in e2c using OpenFeature with multiple provider options.

## Configuration

Feature flags are configured in your e2c config file (typically located at `~/.config/e2c/config.yaml`):

```yaml
feature_flags:
  enabled: true
  provider: "configcat"  # Options: "configcat" or "env"
  
  # ConfigCat provider settings
  configcat:
    sdk_key: "YOUR_CONFIGCAT_SDK_KEY"
    environment: "production"  # Optional: override environment
    base_url: ""               # Optional: for on-premise installations
    cache_ttl_seconds: 60      # Optional: cache TTL in seconds
    polling_interval_seconds: 60 # Optional: polling interval in seconds
  
  # Environment variable provider settings
  env:
    prefix: "E2C_FEATURE_"     # Optional: prefix for environment variables
    case_sensitive: false      # Optional: case sensitivity for env var keys
```

You can also set these values using environment variables:

```bash
export E2C_FEATURE_FLAGS_ENABLED=true
export E2C_FEATURE_FLAGS_PROVIDER=configcat
export E2C_FEATURE_FLAGS_CONFIGCAT_SDK_KEY=your-sdk-key
export E2C_FEATURE_FLAGS_CONFIGCAT_ENVIRONMENT=production
```

### Selecting a Provider via Command Line

You can override the provider configuration using the `--openfeature-provider` flag:

```bash
# Use the environment variable provider
e2c --openfeature-provider=env

# Use the ConfigCat provider
e2c --openfeature-provider=configcat
```

## Using Feature Flags in Code

Import the feature flags package:

```go
import (
    "context"
    "github.com/nlamirault/e2c/internal/featureflags"
)
```

The usage remains the same regardless of which provider you're using. The provider implementation is abstracted away behind the OpenFeature API.

### Boolean Flags

```go
// Using a boolean flag with a default value of false
enabled := featureflags.GetBoolValue(ctx, "my_new_feature", false)
if enabled {
    // The feature is enabled
} else {
    // The feature is disabled
}
```

### String Flags

```go
// Using a string flag with a default value
theme := featureflags.GetStringValue(ctx, "ui_theme", "dark")
```

### Numeric Flags

```go
// Using an integer flag
limit := featureflags.GetIntValue(ctx, "api_rate_limit", 100)

// Using a float flag
threshold := featureflags.GetFloatValue(ctx, "confidence_threshold", 0.95)
```

## Targeting Specific Users

To target specific users with flags, use the `EvaluationContext`:

```go
import (
    "github.com/open-feature/go-sdk/pkg/openfeature"
)

// Create an evaluation context with user attributes
ctx := context.Background()
client := featureflags.GetClient()

// Create evaluation context with user targeting information
evalCtx := openfeature.NewEvaluationContext(
    "user-123",
    map[string]interface{}{
        "email": "user@example.com",
        "groups": []string{"beta-testers", "premium"},
        "region": "us-west",
    },
)

// Evaluate a flag with the context
premium := client.BooleanValue(ctx, "premium_features", false, evalCtx)
```

## Provider-Specific Details

### ConfigCat Provider

The ConfigCat provider connects to the ConfigCat service to retrieve feature flags:

- Requires an SDK key from your ConfigCat account
- Supports targeting rules, percentile rollouts, and A/B testing
- Provides a management UI for configuring flags
- Supports multiple environments (dev, staging, production)

### Environment Variable Provider

The environment variable provider reads flags from environment variables:

- No external service required
- Simple to set up and use
- Environment variables should be named with the pattern: `[PREFIX]_[FLAG_NAME]`
- Values are parsed based on their type:
  - `true`, `1`, `yes`, `on` are parsed as boolean `true`
  - Numbers are parsed as integers or floats
  - All other values are treated as strings

Example environment variables:

```bash
# For a boolean flag named "new_ui"
export E2C_FEATURE_NEW_UI=true

# For a numeric flag named "max_connections"
export E2C_FEATURE_MAX_CONNECTIONS=100

# For a string flag named "theme"
export E2C_FEATURE_THEME=dark
```

## Best Practices

1. Always provide sensible default values that maintain backward compatibility
2. Use descriptive flag names that indicate their purpose
3. Clean up flags that are no longer needed after they've been fully rolled out
4. Consider defining constants for flag names to avoid typos and ensure consistency
5. Use the environment variable provider for local development and testing
6. Use the ConfigCat provider for production environments with dynamic control

## Debugging

To troubleshoot feature flag issues:

1. Verify that feature flags are enabled in the configuration
2. Check that the correct provider is selected
3. For ConfigCat, verify that the SDK key is correct
4. For environment variables, check that they're properly set with the correct prefix
5. Use a higher log level (`--log-level=debug`) to see more detailed information about flag evaluations