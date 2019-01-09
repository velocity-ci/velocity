defmodule ArchitectWeb.Queries.KnownHostsQueries do
  use Absinthe.Schema.Notation
  alias ArchitectWeb.Resolvers

  object :known_hosts_queries do
    @desc "Get all known hosts"
    field :known_hosts, list_of(:known_host) do
      resolve(&Resolvers.KnownHosts.list_known_hosts/3)
    end

    @desc "Get fingerprint for host"
    field :for_host, :known_host do
    end
  end
end
