defmodule ArchitectWeb.Subscriptions.KnownHostSubscriptions do
  use Absinthe.Schema.Notation
  alias Ecto.Changeset
  alias Architect.KnownHosts.KnownHost

  object :known_hosts_subscriptions do
    field :known_host_added, non_null(:known_host) do
      trigger(:for_host,
        topic: fn
          %KnownHost{id: id} ->
            ["all", id]

          %Changeset{} ->
            []
        end
      )

      config(fn _args, _info ->
        {:ok, topic: "all"}
      end)
    end

    field :known_host_verified, non_null(:known_host) do
      trigger(:verify,
        topic: fn
          %KnownHost{id: id} ->
            ["all", id]

          %Changeset{} ->
            []
        end
      )

      config(fn _args, _info ->
        {:ok, topic: "all"}
      end)
    end
  end
end
