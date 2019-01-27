defmodule ArchitectWeb.Queries.ProjectsQueries do
  use Absinthe.Schema.Notation
  alias ArchitectWeb.{Resolvers, Middleware}
  alias Absinthe.Resolution
  alias Architect.Projects
  use Absinthe.Relay.Schema.Notation, :modern

  #  @desc "List projects"
  #  connection field(:projects, node_type: :project) do
  #    middleware(Middleware.Authorize)
  #    resolve(&Resolvers.Projects.list_projects/2)
  #  end
  #
  #  @desc "List commits"
  #  connection field(:commits, node_type: :commit) do
  #    middleware(Middleware.Authorize)
  #
  #    arg(:project_id, non_null(:string))
  #    arg(:branch, non_null(:string))
  #
  #    # Add the project to the context
  #    middleware(fn res, _ ->
  #      [%{argument_data: %{project_id: project_id}} | _] = res.path
  #      project = Projects.get_project!(project_id)
  #      %{res | context: Map.put(res.context, :project, project)}
  #    end)
  #
  #    resolve(&Resolvers.Projects.list_commits/2)
  #  end
end
