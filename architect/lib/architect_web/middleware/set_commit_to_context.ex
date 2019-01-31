defmodule ArchitectWeb.Middleware.SetCommitToContext do
  @moduledoc """
  Middleware that just sets the source to the commit key in context
  """
  alias Git.Commit

  @behaviour Absinthe.Middleware
  def call(%{source: %Commit{} = commit} = res, _) do
    %{res | context: Map.put(res.context, :commit, commit)}
  end
end
