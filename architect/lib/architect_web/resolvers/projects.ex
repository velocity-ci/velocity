defmodule ArchitectWeb.Resolvers.Projects do
  alias Architect.Projects

  def list_projects(_parent, _args, _resolution) do
    {:ok, Projects.list_projects()}
  end

  def list_commits_for_project(project, %{branch: branch}, _resolution) do
    {:ok, Projects.list_commits(project, branch)}
  end

  def list_branches_for_project(project, _args, _resolution) do
    {:ok, Projects.list_branches(project)}
  end

  def get_default_branch_for_project(project, _args, _resoltion) do
    {:ok, Projects.default_branch(project)}
  end
end
