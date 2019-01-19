defmodule ArchitectWeb.Resolvers.Projects do
  alias Architect.Projects

  def list_projects(_parent, _args, _resolution) do
    {:ok, Projects.list_projects()}
  end

  def list_commits_for_project(_parent, %{project_id: project_id, branch: branch}, _resolution) do
    project = Projects.get_project!(project_id)
    {:ok, Projects.list_commits(project, branch)}
  end

  def list_branches_for_project(_parent, %{project_id: project_id}, _resolution) do
    project = Projects.get_project!(project_id)
    {:ok, Projects.list_branches(project)}
  end
end
