defmodule ArchitectWeb.Mutations.KnownHostsMutations do
  use Absinthe.Schema.Notation
  alias Architect.KnownHosts
  require Logger

  object :known_hosts_mutations do
    @desc "Create unverified known host"
    field :create_known_host, non_null(:known_host_payload) do
      middleware(ArchitectWeb.Middleware.Authorize)

      arg(:host, non_null(:string))

      resolve(fn params, %{context: _context} ->
        with {:ok, known_host} <- Architect.KnownHosts.create_known_host(params) do
          {:ok, known_host}
        else
          {:error, %Ecto.Changeset{} = changeset} ->
            {:ok, changeset}

          error ->
            Logger.error("Create error #{inspect(error)}")
            {:error, "Unknown error"}
        end
      end)
    end

    @desc "Verify a known host"
    field :verify_known_host, non_null(:known_host_payload) do
      middleware(ArchitectWeb.Middleware.Authorize)

      arg(:id, non_null(:string))

      resolve(fn %{id: id}, %{context: context} ->
        known_host = KnownHosts.get_known_host!(id)

        with {:ok, known_host} <- KnownHosts.update_known_host(known_host, %{verified: true}) do
          {:ok, known_host}
        else
          {:error, %Ecto.Changeset{} = changeset} ->
            {:ok, changeset}

          error ->
            Logger.error("Verify error #{inspect(error)}")
            {:error, "Unknown error"}
        end
      end)
    end
  end
end
