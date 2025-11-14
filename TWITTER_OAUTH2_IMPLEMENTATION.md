# Twitter (X) OAuth2 Provider - Implementation Summary

## Overview

This implementation adds official support for Twitter (X) OAuth2 authentication to Ory Kratos. The new provider uses Twitter's modern API v2 and OAuth 2.0 protocol, providing a more robust and future-proof integration compared to the legacy OAuth1 provider.

## Files Created/Modified

### New Files

1. **`selfservice/strategy/oidc/provider_twitter_v2.go`**
   - Main implementation of the Twitter OAuth2 provider
   - Implements the `OAuth2Provider` interface
   - Handles OAuth2 authorization flow with PKCE support
   - Fetches user claims from Twitter API v2 `/users/me` endpoint
   - Maps Twitter user data to Kratos claims structure

2. **`selfservice/strategy/oidc/provider_twitter_v2_test.go`**
   - Comprehensive test suite for the Twitter OAuth2 provider
   - Tests provider configuration, OAuth2 setup, and claims handling
   - Verifies provider registration in the supported providers map
   - Validates scope checking and error handling

3. **`TWITTER_OAUTH2_GUIDE.md`**
   - Complete user documentation
   - Configuration examples
   - Migration guide from OAuth1 to OAuth2
   - Troubleshooting section
   - Example implementations

### Modified Files

1. **`selfservice/strategy/oidc/provider_config.go`**
   - Added `twitter_v2` to the `supportedProviders` map
   - Updated provider documentation comments to include Twitter OAuth2

## Technical Implementation Details

### Provider Features

- **Protocol**: OAuth 2.0 with PKCE (Proof Key for Code Exchange)
- **API Version**: Twitter API v2
- **Endpoints**:
  - Authorization: `https://twitter.com/i/oauth2/authorize`
  - Token: `https://api.twitter.com/2/oauth2/token`
  - User Info: `https://api.twitter.com/2/users/me`

### Supported Scopes

The provider supports all Twitter OAuth2 scopes, with common examples:
- `tweet.read` - Read tweets
- `users.read` - Read user profile (required for authentication)
- `offline.access` - Refresh tokens support

### User Claims Mapping

The provider maps Twitter API v2 user data to the following claims:

```go
Claims {
    Issuer:            "https://api.twitter.com/2/oauth2/token"
    Subject:           user.ID
    Name:              user.Name
    Nickname:          user.Username
    PreferredUsername: user.Username
    Picture:           user.ProfileImageURL
    RawClaims: {
        "description": user.Description
        "verified":    user.Verified
    }
}
```

### Security Features

1. **PKCE Support**: Automatically includes `code_challenge_method=S256` in authorization requests
2. **Scope Validation**: Verifies granted scopes match requested scopes
3. **Error Handling**: Comprehensive error handling with descriptive messages
4. **Private Network Protection**: Compatible with Kratos' SSRF protection (uses fixed token URL)

## Usage Example

### Basic Configuration

```yaml
selfservice:
  methods:
    oidc:
      enabled: true
      config:
        providers:
          - id: twitter
            provider: twitter_v2
            client_id: YOUR_CLIENT_ID
            client_secret: YOUR_CLIENT_SECRET
            mapper_url: file:///etc/config/kratos/twitter-mapper.jsonnet
            scope:
              - tweet.read
              - users.read
```

### JSONNet Mapper

```jsonnet
local claims = std.extVar('claims');

{
  identity: {
    traits: {
      email: claims.email,
      name: claims.name,
      username: claims.preferred_username,
      picture: claims.picture,
    },
  },
}
```

## Testing

The implementation includes comprehensive tests:

- ✅ Provider configuration validation
- ✅ OAuth2 config generation
- ✅ PKCE parameter inclusion
- ✅ Claims retrieval and mapping
- ✅ Scope validation
- ✅ Provider registration
- ✅ Error handling

Run tests with:
```bash
go test -v -run TestProviderTwitterV2 ./selfservice/strategy/oidc/
```

## Migration Path

For users currently using the `x` (OAuth1) provider:

1. **Enable OAuth 2.0** in Twitter Developer Portal
2. **Update Kratos configuration** to use `twitter_v2` provider
3. **Update callback URLs** in Twitter App settings
4. **Test in development** environment first
5. **Deploy to production** with monitoring

## Backward Compatibility

- The new `twitter_v2` provider does **not** replace the existing `x` (OAuth1) provider
- Both providers can coexist in the same Kratos instance
- Users can migrate at their own pace
- No breaking changes to existing OAuth1 integrations

## Benefits Over OAuth1

1. **Modern Protocol**: OAuth 2.0 is the industry standard
2. **Better Security**: PKCE provides protection against authorization code interception
3. **Future-Proof**: Twitter is deprecating OAuth 1.0a
4. **Simpler Flow**: No request signing complexity
5. **Better Error Messages**: More detailed error responses from Twitter API v2

## Compliance

The implementation follows Ory Kratos standards:

- ✅ Consistent code style with existing providers
- ✅ Proper error handling with herodot
- ✅ Comprehensive logging
- ✅ Telemetry/tracing support (compatible with existing tracing)
- ✅ Proper interface implementation
- ✅ Apache 2.0 license headers
- ✅ Documentation and examples

## Future Enhancements

Potential improvements for future versions:

1. **Email Claim Support**: Add email extraction from Twitter's email endpoint (requires additional scope)
2. **Organization Support**: Add support for Ory Network organizations
3. **Additional Claims**: Map more Twitter user fields (followers, verified status, etc.)
4. **Refresh Token Support**: Implement refresh token flow with `offline.access` scope

## References

- [Twitter OAuth 2.0 Documentation](https://developer.twitter.com/en/docs/authentication/oauth-2-0)
- [Twitter API v2 User Lookup](https://developer.twitter.com/en/docs/twitter-api/users/lookup/introduction)
- [PKCE RFC 7636](https://datatracker.ietf.org/doc/html/rfc7636)
- [OAuth 2.0 RFC 6749](https://datatracker.ietf.org/doc/html/rfc6749)

## Related GitHub Issues

This implementation addresses:
- #3778 - feat: add twitter SSO
- #2117 - feat: Provider Twitter  
- #2116 - Add Twitter as an identity provider
- #517 - Social sign in with Twitter

## Author

Implementation by: GitHub Copilot (with user guidance)
License: Apache 2.0
Copyright: © 2024 Ory Corp

---

**Status**: ✅ Ready for Review
**Type**: Feature Addition
**Breaking Changes**: None
**Testing**: Comprehensive test coverage included
