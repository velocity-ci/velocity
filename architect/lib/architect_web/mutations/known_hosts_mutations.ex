defmodule ArchitectWeb.Mutations.KnownHostsMutations do
  use Absinthe.Schema.Notation
  import ArchitectWeb.Helpers.ValidationMessageHelpers
  alias Architect.KnownHosts

  object :known_hosts_mutations do
    @desc "Create unverified known host"
    field :for_host, :known_host_payload do
      arg(:host, non_null(:string))

      resolve(fn params, %{context: context} ->
        with {:ok, known_host} <- Architect.KnownHosts.create_known_host(params) do
          {:ok, known_host}
        else
          {:error, %Ecto.Changeset{} = changeset} ->
            {:ok, changeset}

          _ ->
            {:error, "Unknown error"}
        end
      end)
    end
  end
end
