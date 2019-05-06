# Architect

## Development

```
docker-compose up
```

or 

```
scripts/get-v-ssh-keyscan.sh
scripts/get-vcli.sh
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

Dev internal:
```
{:ok, u} = Architect.Accounts.create_user(%{username: "admin", password: "admin"})
or
[u] = Architect.Accounts.list_users()

{:ok, {p, e}} = Architect.Projects.create_project(u, "https://github.com/velocity-ci/velocity.git")
or
p = Architect.Projects.get_project_by_slug!("velocity-ci-velocity-at-github-com")

Architect.Builds.create_build(u, p, "master", "671663fda8cad9545661346017f90eff74a2d248", "examples/hello-velocity")
```

## Tests
```
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```
