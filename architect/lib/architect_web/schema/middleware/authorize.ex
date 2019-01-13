defmodule ArchitectWeb.Schema.Middleware.Authorize do
  @behaviour Absinthe.Middleware

  import ArchitectWeb.Helpers.ValidationMessageHelpers

  def call(resolution, _config) do
    case resolution.context do
      # %{current_user: %User{}} ->
      #   resolution

      _ ->
        Absinthe.Resolution.put_result(resolution, {:ok, generic_message("not authorized")})
    end
  end
end
