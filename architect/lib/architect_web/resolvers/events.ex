defmodule ArchitectWeb.Resolvers.Events do
  alias Absinthe.Resolution
  alias Absinthe.Relay
  alias Architect.Repo
  alias Architect.Events
  alias Architect.Events.Event
  import Ecto.Query

  alias Architect.KnownHosts.KnownHost

  def list_events(args, _res) do
    Events.list_events_query()
    |> Relay.Connection.from_query(&Repo.all/1, args)
  end
end
