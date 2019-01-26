defmodule ArchitectWeb.Queries.ProjectsQueries do
  use Absinthe.Schema.Notation
  alias ArchitectWeb.Resolvers
  alias Absinthe.Resolution
  alias Architect.Projects

  object :projects_queries do
    @desc "List projects"
    field :list_projects, non_null(list_of(non_null(:project))) do
      middleware(ArchitectWeb.Middleware.Authorize)
      resolve(&Resolvers.Projects.list_projects/3)
    end

    @desc "List commits for project"
    field :list_commits, non_null(list_of(non_null(:commit))) do
      arg(:project_id, non_null(:string))
      arg(:branch, non_null(:string))

      middleware(fn res, _ ->
        [%{argument_data: %{project_id: project_id}} | _] = res.path
        project = Projects.get_project!(project_id)
        %{res | context: Map.put(res.context, :project, project)}
      end)

      resolve(fn parent, args, res ->
        task =
          Task.async(fn ->
            {:ok, Resolvers.Projects.list_commits_for_project(parent, args, res)}
          end)

        {:middleware, Elixir.Absinthe.Middleware.Async, task}
      end)
    end
  end
end
