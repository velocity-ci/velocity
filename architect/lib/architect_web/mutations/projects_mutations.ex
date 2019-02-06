defmodule ArchitectWeb.Mutations.ProjectsMutations do
  use Absinthe.Schema.Notation
  alias Architect.Projects
  require Logger

  object :projects_mutations do
    @desc "Create project"
    field :create_project, non_null(:project_payload) do
      #      middleware(ArchitectWeb.Middleware.Authorize)

      arg(:address, non_null(:string))

      resolve(fn %{address: address}, %{context: %{current_user: user}} ->
        with {:ok, project} <- Projects.create_project(user, address) do
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
