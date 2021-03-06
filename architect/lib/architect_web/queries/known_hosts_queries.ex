defmodule ArchitectWeb.Queries.KnownHostsQueries do
  use Absinthe.Schema.Notation
  alias ArchitectWeb.Resolvers

  object :known_hosts_queries do
    @desc "Get all known hosts"
    field :list_known_hosts, non_null(list_of(non_null(:known_host))) do
      middleware(ArchitectWeb.Middleware.Authorize)

      resolve(&Resolvers.KnownHosts.list_known_hosts/3)
    end
  end
end
