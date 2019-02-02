defmodule ArchitectWeb.Resolvers.Projects do
  alias Architect.Projects
  alias Architect.Projects.Project
  alias Git.{Branch, Commit}
  alias Absinthe.Resolution
  alias Absinthe.Relay

  def list_projects(pagination_args, _) do
    Projects.list_projects()
    |> Relay.Connection.from_list(pagination_args)
  end


  def get_project(%{slug: slug}, _) do
    {:ok, Projects.get_project_by_slug!(slug) }
  end


  def list_commits(pagination_args, %{context: %{project: project}}) do
    project
    |> Projects.list_commits(pagination_args.branch)
    |> Relay.Connection.from_list(pagination_args)
  end


  def get_branch(%{branch: branch}, %{context: %{project: project}}) do
    {:ok, Projects.get_branch(project, branch)}

  end

  def list_commits(%Branch{name: branch}, pagination_args, %{context: %{project: project}}) do
    project
    |> Projects.list_commits(branch)
    |> Relay.Connection.from_list(pagination_args)
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

  def list_branches_for_project(pagination_args, %{context: %{project: project}}) do
    project
    |> Projects.list_branches()
    |> Relay.Connection.from_list(pagination_args)
  end

  def get_default_branch_for_project(project, _args, _resolution) do
    {:ok, Projects.default_branch(project)}
  end

  def get_commit_count_for_branch(%Branch{name: branch}, _args, %{context: %{project: project}}) do
    {:ok, Projects.commit_count(project, branch)}
  end
end
