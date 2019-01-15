defmodule ArchitectWeb.Middleware.Authorize do
  @behaviour Absinthe.Middleware
  alias Architect.Accounts.User
  alias Kronky.ValidationMessage

  def call(resolution, _config) do
    case resolution.context do
      %{current_user: %User{}} ->
        resolution

      _ ->
        Absinthe.Resolution.put_result(resolution, {:error, "Unauthorized"})
    end
  end
end
