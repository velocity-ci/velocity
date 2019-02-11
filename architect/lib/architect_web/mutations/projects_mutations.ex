defmodule ArchitectWeb.Mutations.ProjectsMutations do
  use Absinthe.Schema.Notation
  alias Architect.Projects
  require Logger
  alias Absinthe.Subscription
  alias Architect.Repo

  object :projects_mutations do
    @desc "Create project"
    field :create_project, non_null(:project_payload) do
      #      middleware(ArchitectWeb.Middleware.Authorize)

      arg(:address, non_null(:string))

      resolve(fn %{address: address}, %{context: %{current_user: user}} ->
        with {:ok, {project, event}} <- Projects.create_project(user, address) do
          Task.async(fn ->
            event = Repo.preload(event, [:user, :project, :known_host])
            Subscription.publish(ArchitectWeb.Endpoint, event, event_added: "all")
          end)

          {:ok, project}
        else
          {:error, %Ecto.Changeset{} = changeset} ->
            {:ok, changeset}

          error ->
            Logger.error("Create error #{inspect(error)}")
            {:error, "Unknown error"}
        end
      end)
    end
  end
end
