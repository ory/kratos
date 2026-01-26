# Use the official Ory Kratos image pinned to a secure digest
FROM oryd/kratos@sha256:e8014c6c58b68e9d8bea4160d3e271b0d05b3db221af379afb6798b603e88ee9

# Copy your configuration files into the container
COPY ./config/kratos.yml /etc/config/kratos/kratos.yml
COPY ./config/identity.schema.json /etc/config/kratos/identity.schema.json

# Run Kratos with your configuration
CMD ["serve", "-c", "/etc/config/kratos/kratos.yml"]
