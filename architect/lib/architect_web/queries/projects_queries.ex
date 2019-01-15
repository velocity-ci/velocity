defmodule ArchitectWeb.Queries.ProjectsQueries do
  use Absinthe.Schema.Notation
  alias ArchitectWeb.Resolvers

  object :projects_queries do
    @desc "List projects"
    field :list_projects, list_of(:project) do
      middleware(ArchitectWeb.Middleware.Authorize)

      resolve(&Resolvers.Projects.list_projects/3)
    end
  end
end
