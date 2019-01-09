defmodule ArchitectWeb.Resolvers.Projects do
  def list_projects(_parent, _args, _resolution) do
    {:ok, Architect.Projects.list_projects()}
  end
end
