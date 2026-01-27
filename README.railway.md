# Deploy Ory Kratos on Railway

This guide helps you deploy Ory Kratos identity and user management system on Railway with automatic database migrations.

## Quick Deploy

Deploy Ory Kratos to Railway with one click:

[![Deploy on Railway](https://railway.app/button.svg)](https://railway.app/template/kratos)

This will automatically:
- 🚀 Deploy the Kratos service
- 🗄️ Provision a PostgreSQL database
- 🔄 Run database migrations automatically
- 🔐 Generate secure secrets
- 🌐 Create a public domain

## Manual Deployment

### Prerequisites

- A [Railway account](https://railway.app/)
- Railway CLI installed (optional): `npm i -g @railway/cli`

### Step 1: Create a New Project

1. Go to [Railway](https://railway.app/)
2. Click "New Project"
3. Choose "Deploy from GitHub repo"
4. Select this repository

### Step 2: Add PostgreSQL Database

1. In your Railway project, click "New"
2. Select "Database"
3. Choose "PostgreSQL"
4. Railway will automatically provide these environment variables:
   - `POSTGRES_HOST`
   - `POSTGRES_PORT`
   - `POSTGRES_USER`
   - `POSTGRES_PASSWORD`
   - `POSTGRES_DATABASE`

### Step 3: Configure Environment Variables

Add the following environment variables to your Kratos service:

#### Required Secrets

Generate secure secrets using OpenSSL:

```bash
# Generate COOKIE_SECRET (32 bytes hex = 64 characters)
openssl rand -hex 32

# Generate CIPHER_SECRET (32 bytes hex = 64 characters)
openssl rand -hex 32
```

Add these to Railway:
- `COOKIE_SECRET`: Your generated cookie secret
- `CIPHER_SECRET`: Your generated cipher secret

#### Automatic Variables

Railway automatically provides:
- `RAILWAY_STATIC_URL`: Your public domain (e.g., `kratos-production.up.railway.app`)

The entrypoint script automatically constructs the database DSN from the PostgreSQL variables.

### Step 4: Deploy

1. Railway will automatically build and deploy your service
2. The entrypoint script will:
   - Wait for PostgreSQL to be ready
   - Run database migrations automatically
   - Start the Kratos server
3. Your Kratos instance will be available at your Railway domain

## Service Endpoints

Once deployed, you'll have access to:

- **Public API**: `https://${RAILWAY_STATIC_URL}:4433` (main API for client applications)
- **Admin API**: `https://${RAILWAY_STATIC_URL}:4434` (internal admin operations)
- **Health Check**: `https://${RAILWAY_STATIC_URL}/health/ready`

## Configuration

### Database Connection

The entrypoint script automatically constructs the PostgreSQL DSN from Railway's environment variables:

```bash
postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DATABASE}?sslmode=disable
```

If you prefer to use a custom `DATABASE_URL`, the script will use that instead.

### Identity Schema

The default identity schema (`/etc/config/kratos/identity.schema.json`) supports:
- Email and password authentication
- Email verification
- Password recovery
- User profile with first and last name

You can customize this by modifying `config/identity.schema.json` in your repository.

### Authentication Methods

Enabled by default:
- ✅ Password authentication
- ✅ TOTP (Time-based One-Time Password)
- ✅ Recovery codes (lookup secrets)
- ✅ Email verification codes
- ✅ Password recovery codes

### Self-Service Flows

All self-service flows are pre-configured:
- Registration: `https://${RAILWAY_STATIC_URL}/registration`
- Login: `https://${RAILWAY_STATIC_URL}/login`
- Recovery: `https://${RAILWAY_STATIC_URL}/recovery`
- Verification: `https://${RAILWAY_STATIC_URL}/verification`
- Settings: `https://${RAILWAY_STATIC_URL}/settings`
- Error: `https://${RAILWAY_STATIC_URL}/error`

**Note**: These URLs expect a frontend application. You'll need to build UI pages for these flows or use Ory's hosted UI.

## Automatic Migrations

The entrypoint script (`entrypoint.sh`) handles database migrations automatically:

1. **Waits for PostgreSQL**: Retries connection up to 30 times (60 seconds total)
2. **Shows current status**: Displays migration status before running
3. **Runs migrations**: Executes `kratos migrate sql up -e --yes`
4. **Confirms success**: Shows migration status after completion
5. **Starts server**: Launches Kratos with `serve all` command

### Migration Logs

Check your Railway deployment logs to see migration progress:

```
🚀 Starting Ory Kratos setup...
✅ Using Railway PostgreSQL
📊 Database connection configured
⏳ Waiting for PostgreSQL to be ready...
✅ Database is ready!
📋 Current migration status:
...
🔄 Running database migrations...
...
🎉 Migrations completed successfully!
🚀 Starting Kratos server...
```

## Troubleshooting

### Database Connection Issues

If migrations fail:

1. Check PostgreSQL service is running in Railway
2. Verify environment variables are set correctly
3. Check Railway logs for connection errors
4. Ensure PostgreSQL service is linked to Kratos service

### Secret Generation

Always use strong secrets:

```bash
# Good: 32 bytes = 64 hex characters
openssl rand -hex 32

# Bad: Short or predictable secrets
# Don't use: "mysecret123" or other simple strings
```

### Health Check Failing

The health check endpoint is `/health/ready`. If it fails:

1. Check Kratos logs in Railway
2. Verify database migrations completed successfully
3. Ensure both ports 4433 and 4434 are accessible
4. Check that `RAILWAY_STATIC_URL` is set correctly

### Viewing Logs

In Railway:
1. Open your project
2. Click on the Kratos service
3. Go to the "Deployments" tab
4. Click on the latest deployment
5. View logs in real-time

## Security Considerations

1. **Secrets**: Railway automatically generates secure secrets. Store them safely.
2. **HTTPS**: Railway provides HTTPS by default for your domain.
3. **Database**: PostgreSQL runs in a private network, not publicly accessible.
4. **CORS**: CORS is enabled. Configure `allowed_return_urls` in `config/kratos.yml` for your specific domains.
5. **Admin API**: The admin API (port 4434) should be restricted to internal services only.

## Updating Configuration

To update Kratos configuration:

1. Modify `config/kratos.yml` in your repository
2. Commit and push changes
3. Railway will automatically redeploy
4. Migrations will run again (safely, they're idempotent)

## Support

- [Ory Kratos Documentation](https://www.ory.sh/docs/kratos)
- [Railway Documentation](https://docs.railway.app/)
- [Ory Community](https://github.com/ory/kratos/discussions)

## Next Steps

After deployment:

1. **Build a UI**: Create frontend pages for the self-service flows
2. **Configure Email**: Set up SMTP for email verification and recovery
3. **Add OAuth**: Configure social login providers if needed
4. **Customize Schema**: Modify identity schema for your use case
5. **Set up Monitoring**: Use Railway's monitoring features

## License

This deployment configuration is part of the Ory Kratos project. See the main repository for license information.
