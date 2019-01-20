defmodule ArchitectWeb.Subscriptions.ProjectsSubscriptions do
  use Absinthe.Schema.Notation
  alias Ecto.Changeset
  alias Architect.Projects.Project
  require Logger

  object :projects_subscriptions do
    field :project_added, non_null(:project) do
      trigger(:create_project,
        topic: fn
          %Project{id: id} ->
            Logger.debug("Successful createProject subscription")
            ["all", id]

          %Changeset{} ->
            Logger.debug("Failure createProject subscription")
            []
        end
      )

      config(fn _args, _info ->
        {:ok, topic: "all"}
      end)
    end
  end
end
