defmodule ArchitectWeb.Resolvers.KnownHosts do
  def list_known_hosts(_parent, _args, _resolution) do
    {:ok, Architect.KnownHosts.list_known_hosts()}
  end
end
