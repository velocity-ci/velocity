defmodule ArchitectWeb.Queries.ProjectsQueries do
  use Absinthe.Schema.Notation
  alias ArchitectWeb.Resolvers

  object :projects_queries do
    @desc "Get all projects"
    field :projects, list_of(:project) do
      resolve(&Resolvers.Projects.list_projects/3)
    end
  end
end
