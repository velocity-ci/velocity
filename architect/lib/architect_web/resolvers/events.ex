defmodule ArchitectWeb.Resolvers.Events do
  alias Absinthe.Resolution
  alias Absinthe.Relay
  alias Architect.Repo
  alias Architect.Events.Event
  import Ecto.Query

  alias Architect.KnownHosts.KnownHost

  def list_events(args, _res) do
    query =
      from(event in Event,
        left_join: user in assoc(event, :user),
        left_join: known_host in assoc(event, :known_host),
        left_join: project in assoc(event, :project),
        preload: [user: user, known_host: known_host, project: project]
      )

    Relay.Connection.from_query(query, &Repo.all/1, args)
  end
end
