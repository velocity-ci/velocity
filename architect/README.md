# Architect

## Development

```
docker-compose up
```

or

```
scripts/get-v-ssh-keyscan.sh # TODO: use normal ssh-keyscan bin
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
iex -S mix phx.server

{:ok, u} = Architect.Accounts.create_user(%{username: "admin", password: "admin"})

# HTTPS
{:ok, {p, e}} = Architect.Projects.create_project(u, "https://github.com/velocity-ci/velocity.git")

# GIT+SSH
{:ok, {kh, _e}} = Architect.KnownHosts.create_known_host(u, "github.com")
Architect.KnownHosts.verify_known_host(u, kh)
pk = """
"""
{:ok, {p, e}} = Architect.Projects.create_project(u, "git@github.com:velocity-ci/velocity.git", pk)

# or, if already created

[u] = Architect.Accounts.list_users()
p = Architect.Projects.get_project_by_slug!("velocity-ci-velocity-at-github-com")

kh = Architect.KnownHosts.get_known_host_by_host!("github.com")
Architect.KnownHosts.verify_known_host(u, kh)

# create a build
Architect.Builds.create_build(u, p, "master", "23d4e47ee635727a4fb65fea6e1cf1749861c079", "examples/hello-velocity")
```

## Tests

```
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

Blueprint/Task
Pipeline/Tasks

Architect scheduling:

- We have a list of tasks to run which it distributes to builders
- In a pipeline, tasks will only be scheduled once preceeding tasks have finished
