defmodule ArchitectWeb.Subscriptions.ProjectsSubscriptions do
  use Absinthe.Schema.Notation
  alias Ecto.Changeset
  alias Architect.Projects.Project

  object :projects_subscriptions do
    field :project_added, non_null(:project) do
      trigger(:create_project,
        topic: fn
          %Project{id: id} ->
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
