defmodule ArchitectWeb.Resolvers.Users do
  def list_users(_parent, _args, _resolution) do
    {:ok, Architect.Users.list_users()}
  end
end
