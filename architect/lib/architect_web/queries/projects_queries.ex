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
    field :list_commits, list_of(:commit) do
      arg(:project_id, :string)
      arg(:branch, :string)

      resolve(&Resolvers.Projects.list_commits_for_project/3)
    end

    @desc "List branches for project"
    field :list_branches, list_of(:branch) do
      arg(:project_id, :string)

      resolve(&Resolvers.Projects.list_branches_for_project/3)
    end
  end
end
