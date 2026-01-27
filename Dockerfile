# Use the official Ory Kratos image pinned to a secure digest
FROM oryd/kratos@sha256:e8014c6c58b68e9d8bea4160d3e271b0d05b3db221af379afb6798b603e88ee9

# Switch to root to modify file permissions
USER root

# Copy your configuration files into the container
COPY ./config/kratos.yml /etc/config/kratos/kratos.yml
COPY ./config/identity.schema.json /etc/config/kratos/identity.schema.json

# Copy and setup entrypoint script for Railway deployment
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Switch back to ory user for security
USER ory

# Use entrypoint script that handles migrations and starts Kratos
ENTRYPOINT ["/entrypoint.sh"]
