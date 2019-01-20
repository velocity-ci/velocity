defmodule ArchitectWeb.Queries.ProjectsQueries do
  use Absinthe.Schema.Notation
  alias ArchitectWeb.Resolvers

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

      resolve(&Resolvers.Projects.list_commits_for_project/3)
    end
  end
end
