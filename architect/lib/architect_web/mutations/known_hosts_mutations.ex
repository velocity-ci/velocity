defmodule ArchitectWeb.Mutations.KnownHostsMutations do
  use Absinthe.Schema.Notation
  alias Absinthe.Subscription
  alias Architect.KnownHosts
  require Logger
  alias Architect.Repo

  object :known_hosts_mutations do
    @desc "Create unverified known host"
    field :create_known_host, non_null(:known_host_payload) do
      middleware(ArchitectWeb.Middleware.Authorize)

      arg(:host, non_null(:string))

      resolve(fn %{host: host}, %{context: %{current_user: user}} ->
        with {:ok, {known_host, event}} <- KnownHosts.create_known_host(user, host) do
          Task.async(fn ->
            event = Repo.preload(event, [:user, :project, :known_host])
            Subscription.publish(ArchitectWeb.Endpoint, event, event_added: "all")
          end)

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

      resolve(fn %{id: id}, %{context: %{current_user: user}} ->
        known_host = KnownHosts.get_known_host!(id)

        with {:ok, {known_host, event}} <- KnownHosts.verify_known_host(user, known_host) do
          Task.async(fn ->
            event = Repo.preload(event, [:user, :project, :known_host])
            Subscription.publish(ArchitectWeb.Endpoint, event, event_added: "all")
          end)

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
