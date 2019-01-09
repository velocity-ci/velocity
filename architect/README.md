# Architect

## Development

```
docker-compose up
```

or 

```
scripts/get-v-ssh-keyscan.sh
docker-compose up -d db
mix ecto.drop
mix ecto.create
mix ecto.migrate

# start server
mix phx.server

# run tests
mix test --trace
```
If you wish to use your host elixir/otp.

## Tests
```
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```
