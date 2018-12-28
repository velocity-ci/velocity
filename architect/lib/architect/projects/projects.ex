defmodule Architect.Projects do
  @moduledoc """
  The Projects context.
  """

  import Ecto.Query, warn: false
  alias Architect.Repo

  alias Architect.Projects.Project

  @doc """
  Returns the list of projects.

  ## Examples

      iex> list_projects()
      [%Project{}, ...]

  """
  def list_projects() do
    Repo.all(Project)
  end
end
