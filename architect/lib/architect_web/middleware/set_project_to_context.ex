defmodule ArchitectWeb.Middleware.SetProjectToContext do
  @moduledoc """
  Middleware that just sets the source to the project key in context
  """
  alias Architect.Projects.Project

  @behaviour Absinthe.Middleware
  def call(%{source: %Project{} = project} = res, _) do
    %{res | context: Map.put(res.context, :project, project)}
  end
end
