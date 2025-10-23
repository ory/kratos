# Twitter (X) OAuth2 Provider Guide

This guide explains how to use the new Twitter (X) OAuth2 provider in Ory Kratos.

## Overview

The `twitter_v2` provider implements OAuth2 authentication using Twitter's modern API v2. This is the recommended approach for new integrations, as Twitter's OAuth1 protocol is being deprecated.

## Key Features

- ✅ OAuth2 Authorization Code Flow with PKCE
- ✅ Twitter API v2 user endpoints
- ✅ Support for user profile information
- ✅ Configurable scopes
- ✅ Modern and maintained authentication flow

## Differences from OAuth1 Provider (`x`)

| Feature | `x` (OAuth1) | `twitter_v2` (OAuth2) |
|---------|--------------|----------------------|
| Protocol | OAuth 1.0a | OAuth 2.0 with PKCE |
| Twitter API | v1.1 | v2 |
| Status | Deprecated by Twitter | Recommended |
| Email Access | Via `email` scope | Via Twitter App configuration |
| Security | Request signing | Bearer tokens with PKCE |

## Prerequisites

1. **Twitter Developer Account**: Sign up at [developer.twitter.com](https://developer.twitter.com)
2. **Create a Twitter App**: 
   - Go to the [Twitter Developer Portal](https://developer.twitter.com/en/portal/dashboard)
   - Create a new app or use an existing one
   - Enable OAuth 2.0 authentication
3. **Configure OAuth 2.0 Settings**:
   - Set the callback URL to: `https://your-domain.com/self-service/methods/oidc/callback/twitter`
   - Note your **Client ID** and **Client Secret**
   - Configure required scopes (see below)

## Configuration

### Basic Kratos Configuration

Add the following to your Kratos configuration file:

```yaml
selfservice:
  methods:
    oidc:
      enabled: true
      config:
        providers:
          - id: twitter
            provider: twitter_v2
            client_id: YOUR_TWITTER_CLIENT_ID
            client_secret: YOUR_TWITTER_CLIENT_SECRET
            mapper_url: file:///path/to/twitter-mapper.jsonnet
            scope:
              - tweet.read
              - users.read
```

### Available Scopes

Twitter API v2 uses granular scopes. Common scopes include:

- `tweet.read` - Read tweets
- `tweet.write` - Write tweets
- `users.read` - Read user profile information (required for authentication)
- `follows.read` - Read follower information
- `offline.access` - Get refresh tokens (if you need long-lived sessions)

**Note**: The `users.read` scope is typically required for authentication as it allows reading the authenticated user's profile.

### JSONNet Mapper Example

Create a mapper file (`twitter-mapper.jsonnet`) to transform Twitter user data to Kratos identity traits:

```jsonnet
local claims = std.extVar('claims');

{
  identity: {
    traits: {
      [if 'email' in claims then 'email' else null]: claims.email,
      name: {
        first: claims.name,
      },
      username: claims.preferred_username,
      [if 'picture' in claims then 'picture' else null]: claims.picture,
    },
  },
}
```

### Advanced Configuration with PKCE

The Twitter OAuth2 provider automatically uses PKCE (Proof Key for Code Exchange) for enhanced security. This is recommended and enabled by default.

If you need to force PKCE mode in your configuration:

```yaml
selfservice:
  methods:
    oidc:
      config:
        providers:
          - id: twitter
            provider: twitter_v2
            client_id: YOUR_TWITTER_CLIENT_ID
            client_secret: YOUR_TWITTER_CLIENT_SECRET
            pkce: force  # Optional: force, auto, or never
            scope:
              - tweet.read
              - users.read
```

## User Claims

The Twitter OAuth2 provider returns the following claims from the Twitter API v2:

| Claim Field | Description | Twitter Field |
|-------------|-------------|---------------|
| `subject` | Unique user ID | `id` |
| `name` | Display name | `name` |
| `nickname` | Username | `username` |
| `preferred_username` | Username | `username` |
| `picture` | Profile image URL | `profile_image_url` |
| `raw_claims.description` | User bio | `description` |
| `raw_claims.verified` | Verified status | `verified` |

## Migration from OAuth1 (`x` provider)

If you're currently using the `x` (OAuth1) provider, here's how to migrate:

### Step 1: Update Your Twitter App

1. Enable OAuth 2.0 in your Twitter App settings
2. Update callback URLs to use the OAuth2 flow
3. Generate new OAuth 2.0 credentials (Client ID and Secret)

### Step 2: Update Kratos Configuration

Change your provider configuration from:

```yaml
providers:
  - id: twitter
    provider: x  # Old OAuth1
    client_id: OAUTH1_CONSUMER_KEY
    client_secret: OAUTH1_CONSUMER_SECRET
```

To:

```yaml
providers:
  - id: twitter
    provider: twitter_v2  # New OAuth2
    client_id: OAUTH2_CLIENT_ID
    client_secret: OAUTH2_CLIENT_SECRET
    scope:
      - tweet.read
      - users.read
```

### Step 3: Test the Integration

Before rolling out to production:
1. Test the authentication flow in a development environment
2. Verify that user claims are mapped correctly
3. Ensure existing users can still authenticate (if maintaining the same `id`)

## Troubleshooting

### Common Issues

**Issue**: "Scope missing" error
- **Solution**: Ensure you've configured the required scopes in both Kratos config and Twitter App settings

**Issue**: "Invalid redirect URI"
- **Solution**: Verify the callback URL in Twitter Developer Portal matches: `https://your-domain.com/self-service/methods/oidc/callback/twitter`

**Issue**: "User data not returned"
- **Solution**: Check that your Twitter App has the necessary permissions and that you're requesting the `users.read` scope

### Debug Mode

Enable debug logging in Kratos to see detailed OAuth2 flow information:

```yaml
log:
  level: debug
```

## Security Considerations

1. **PKCE**: The provider uses PKCE by default for enhanced security against authorization code interception attacks
2. **HTTPS Required**: Always use HTTPS for your callback URLs in production
3. **Client Secret**: Keep your Twitter Client Secret secure and never commit it to version control
4. **Scope Minimization**: Only request the scopes your application actually needs

## Example Implementation

Here's a complete example with Kratos configuration and a React login button:

### Kratos Configuration (`kratos.yml`)

```yaml
selfservice:
  default_browser_return_url: https://your-app.com/
  flows:
    login:
      ui_url: https://your-app.com/login

  methods:
    oidc:
      enabled: true
      config:
        providers:
          - id: twitter
            provider: twitter_v2
            label: Sign in with Twitter
            client_id: YOUR_TWITTER_CLIENT_ID
            client_secret: YOUR_TWITTER_CLIENT_SECRET
            mapper_url: file:///etc/config/kratos/twitter-mapper.jsonnet
            scope:
              - tweet.read
              - users.read
```

### React Login Button

```jsx
import { useState, useEffect } from 'react'
import { Configuration, V0alpha2Api } from '@ory/client'

const kratos = new V0alpha2Api(
  new Configuration({
    basePath: 'https://your-kratos-instance.com',
    baseOptions: {
      withCredentials: true,
    },
  })
)

function LoginPage() {
  const [flow, setFlow] = useState(null)

  useEffect(() => {
    kratos
      .initializeSelfServiceLoginFlowForBrowsers()
      .then(({ data }) => setFlow(data))
  }, [])

  const twitterProvider = flow?.ui.nodes.find(
    (node) => node.attributes.provider === 'twitter'
  )

  if (!twitterProvider) return <div>Loading...</div>

  return (
    <form action={twitterProvider.attributes.action} method="POST">
      <input
        type="hidden"
        name="provider"
        value={twitterProvider.attributes.value}
      />
      <button type="submit">
        Sign in with Twitter
      </button>
    </form>
  )
}
```

## Additional Resources

- [Twitter OAuth 2.0 Documentation](https://developer.twitter.com/en/docs/authentication/oauth-2-0)
- [Twitter API v2 User Lookup](https://developer.twitter.com/en/docs/twitter-api/users/lookup/introduction)
- [Ory Kratos Social Sign-In Documentation](https://www.ory.sh/docs/kratos/social-signin/overview)
- [Ory Kratos JSONNet Mapper Guide](https://www.ory.sh/docs/kratos/social-signin/data-mapping)

## Contributing

If you encounter issues or have suggestions for improving the Twitter OAuth2 provider:

1. Check existing issues at [github.com/ory/kratos/issues](https://github.com/ory/kratos/issues)
2. Open a new issue with detailed information about your use case
3. Submit a pull request with improvements or bug fixes

## Related Issues

This implementation addresses the following GitHub issues:
- #3778 - feat: add twitter SSO
- #2117 - feat: Provider Twitter
- #2116 - Add Twitter as an identity provider
- #517 - Social sign in with Twitter

---

**License**: Apache 2.0
**Maintainer**: Ory Corp
**Status**: Production Ready
