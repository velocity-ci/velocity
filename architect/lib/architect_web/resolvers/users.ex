defmodule ArchitectWeb.Resolvers.Users do
  def list_users(_parent, _args, _resolution) do
    {:ok, Architect.Accounts.list_users()}
  end
end
