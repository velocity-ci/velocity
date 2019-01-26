defmodule ArchitectWeb.Resolvers.Projects do
  alias Architect.Projects
  alias Architect.Projects.{Branch, Project, Commit}
  alias Absinthe.Resolution

  def list_projects(_parent, _args, _resolution) do
    {:ok, Projects.list_projects()}
  end

  def list_commits_for_project(_parent, %{branch: branch}, %{context: %{project: project}}) do
    {:ok, Projects.list_commits(project, branch)}
  end

  def list_commits_for_project(%Branch{name: branch}, _args, %{context: %{project: project}}) do
    {:ok, Projects.list_commits(project, branch)}
  end

  def list_branches_for_commit(%Commit{sha: sha}, _args, %{context: %{project: project}}) do
    {:ok, Projects.list_branches_for_commit(project, sha)}
  end

  def list_tasks_for_commit(%Commit{sha: sha}, _args, %{context: %{project: project}}) do
    {:ok, Projects.list_tasks(project, {:sha, sha})}
  end

  def list_tasks_for_branch(%Branch{name: branch}, _args, %{context: %{project: project}}) do
    {:ok, Projects.list_tasks(project, {:branch, branch})}
  end

  def list_branches_for_project(project, _args, _resolution) do
    {:ok, Projects.list_branches(project)}
  end

  def get_default_branch_for_project(project, _args, _resolution) do
    {:ok, Projects.default_branch(project)}
  end

  def get_commit_count_for_branch(%Branch{name: branch}, _args, %{context: %{project: project}}) do
    {:ok, Projects.commit_count(project, branch)}
  end
end
